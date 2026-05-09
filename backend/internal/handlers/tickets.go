package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"meddoc/internal/middleware"
	"meddoc/internal/models"
)

// ─── User Ticket Handlers ────────────────────────────────────────────────────

func (h *Handler) ListTickets(c *gin.Context) {
	id := middleware.GetUserID(c)
	tickets, err := h.tickets.ListByUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (h *Handler) CreateTicket(c *gin.Context) {
	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t := &models.Ticket{
		UserID:          middleware.GetUserID(c),
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Phone:           req.Phone,
		Position:        req.Position,
		Room:            req.Room,
		Division:        req.Division,
		Description:     req.Description,
		InventoryNumber: req.InventoryNumber,
		IPAddress:       req.IPAddress,
		Priority:        req.Priority,
		Status:          models.StatusOpen,
	}

	if err := h.tickets.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	// Users can only see their own tickets
	if t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) UpdateTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if t.Status == models.StatusClosed || t.Status == models.StatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit closed/cancelled ticket"})
		return
	}

	var req models.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.tickets.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) CancelTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	h.tickets.SetStatus(c.Request.Context(), id, models.StatusCancelled)
	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}

func (h *Handler) CloseTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	h.tickets.SetStatus(c.Request.Context(), id, models.StatusClosed)
	c.JSON(http.StatusOK, gin.H{"message": "closed"})
}

// ─── Admin Ticket Handlers ───────────────────────────────────────────────────

func (h *Handler) AdminListTickets(c *gin.Context) {
	f := models.TicketFilter{
		Division:  c.Query("division"),
		Priority:  c.Query("priority"),
		Status:    c.Query("status"),
		DateFrom:  c.Query("date_from"),
		DateTo:    c.Query("date_to"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}
	tickets, err := h.tickets.ListAll(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (h *Handler) AdminDeleteTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.tickets.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) AdminCloseTicket(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.tickets.SetStatus(c.Request.Context(), id, models.StatusClosed)
	c.JSON(http.StatusOK, gin.H{"message": "closed"})
}

func (h *Handler) AddComment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req models.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.tickets.AddComment(c.Request.Context(), id, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment added"})
}

func (h *Handler) GetIPLogs(c *gin.Context) {
	logs, err := h.admins.GetIPLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// ─── Export ──────────────────────────────────────────────────────────────────

func (h *Handler) ExportTickets(c *gin.Context) {
	format := c.Query("format") // xml | xlsx
	f := models.TicketFilter{
		Division: c.Query("division"),
		Priority: c.Query("priority"),
		Status:   c.Query("status"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}
	tickets, err := h.tickets.ListAll(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch format {
	case "xml":
		exportXML(c, tickets)
	default:
		// For XLSX/DOCX: return JSON and let frontend handle download
		// In production add unioffice/excelize here
		c.JSON(http.StatusOK, gin.H{
			"format":  format,
			"tickets": tickets,
			"message": "use frontend export library for docx/xlsx",
		})
	}
}

// ─── XML Export ──────────────────────────────────────────────────────────────

type xmlTickets struct {
	XMLName xml.Name     `xml:"Tickets"`
	Items   []xmlTicket  `xml:"Ticket"`
}

type xmlTicket struct {
	ID          int64  `xml:"ID"`
	FirstName   string `xml:"FirstName"`
	LastName    string `xml:"LastName"`
	Division    string `xml:"Division"`
	Description string `xml:"Description"`
	Priority    string `xml:"Priority"`
	Status      string `xml:"Status"`
	CreatedAt   string `xml:"CreatedAt"`
}

func exportXML(c *gin.Context, tickets []*models.Ticket) {
	data := xmlTickets{}
	for _, t := range tickets {
		data.Items = append(data.Items, xmlTicket{
			ID:          t.ID,
			FirstName:   t.FirstName,
			LastName:    t.LastName,
			Division:    t.Division,
			Description: t.Description,
			Priority:    string(t.Priority),
			Status:      string(t.Status),
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	out, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="tickets.xml"`))
	c.Data(http.StatusOK, "application/xml", append([]byte(xml.Header), out...))
}
