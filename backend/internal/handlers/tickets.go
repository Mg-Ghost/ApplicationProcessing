package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"meddoc/internal/auth"
	"meddoc/internal/middleware"
	"meddoc/internal/models"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func parseID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}

// loadMessages загружает сообщения и кладёт в ticket.Messages
func (h *Handler) loadMessages(c *gin.Context, t *models.Ticket) {
	msgs, err := h.messages.ListByTicket(c.Request.Context(), t.ID)
	if err == nil {
		t.Messages = msgs
	}
}

// ─── User: список и просмотр ──────────────────────────────────────────────────

func (h *Handler) ListTickets(c *gin.Context) {
	id := middleware.GetUserID(c)
	tickets, err := h.tickets.ListByUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (h *Handler) GetTicket(c *gin.Context) {
	id := parseID(c)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	if t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	// Помечаем сообщения от админа как прочитанные пользователем
	h.messages.MarkReadByUser(c.Request.Context(), id)
	h.loadMessages(c, t)
	c.JSON(http.StatusOK, t)
}

// ─── User: создание ───────────────────────────────────────────────────────────

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

// ─── User: редактирование полей заявки ───────────────────────────────────────

func (h *Handler) UpdateTicket(c *gin.Context) {
	id := parseID(c)
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

// ─── User: ответное сообщение в чате заявки ───────────────────────────────────

func (h *Handler) UserReply(c *gin.Context) {
	id := parseID(c)

	// Проверяем что заявка принадлежит пользователю
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if t.Status == models.StatusClosed || t.Status == models.StatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "заявление закрыто, переписка недоступна"})
		return
	}

	var req models.AddUserReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем имя пользователя для отображения
	userID := middleware.GetUserID(c)
	u, err := h.users.GetByID(c.Request.Context(), userID)
	authorName := "Пользователь"
	if err == nil {
		authorName = u.FirstName + " " + u.LastName
	}

	msg := &models.TicketMessage{
		TicketID:   id,
		Author:     models.AuthorUser,
		AuthorName: authorName,
		Text:       req.Text,
	}
	if err := h.messages.Add(c.Request.Context(), msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// ─── User: отмена и закрытие ──────────────────────────────────────────────────

func (h *Handler) CancelTicket(c *gin.Context) {
	id := parseID(c)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	h.tickets.SetStatus(c.Request.Context(), id, models.StatusCancelled)
	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}

func (h *Handler) CloseTicket(c *gin.Context) {
	id := parseID(c)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil || t.UserID != middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	userID := middleware.GetUserID(c)
	u, uerr := h.users.GetByID(c.Request.Context(), userID)
	closedBy := "Пользователь"
	if uerr == nil {
		closedBy = u.FirstName + " " + u.LastName
	}
	h.tickets.CloseByAdmin(c.Request.Context(), id, closedBy, "user")
	c.JSON(http.StatusOK, gin.H{"message": "closed"})
}

// ─── Admin: список всех заявок ────────────────────────────────────────────────

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

// ─── Admin: получить заявку с перепиской ──────────────────────────────────────

func (h *Handler) AdminGetTicket(c *gin.Context) {
	id := parseID(c)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Помечаем сообщения от пользователя как прочитанные админом
	h.messages.MarkReadByAdmin(c.Request.Context(), id)
	h.loadMessages(c, t)
	c.JSON(http.StatusOK, t)
}

// UnreadCounts — возвращает количество непрочитанных по каждой заявке
func (h *Handler) UnreadCounts(c *gin.Context) {
	role, _ := c.Get(middleware.ContextRole)
	if role == "admin" {
		counts, err := h.messages.UnreadPerTicketForAdmin(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, counts)
	} else {
		userID := middleware.GetUserID(c)
		counts, err := h.messages.UnreadPerTicketForUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, counts)
	}
}

// ─── Admin: ответ в переписке ─────────────────────────────────────────────────

func (h *Handler) AddComment(c *gin.Context) {
	id := parseID(c)
	t, err := h.tickets.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req models.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminName := getAdminName(c)

	// Обновляем поле admin_comment
	if err := h.tickets.AddComment(c.Request.Context(), id, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем как сообщение в переписке с именем администратора
	msg := &models.TicketMessage{
		TicketID:    id,
		Author:      models.AuthorAdmin,
		AuthorName:  adminName,
		Text:        req.Comment,
		ReadByUser:  false,
		ReadByAdmin: true,
	}
	h.messages.Add(c.Request.Context(), msg)

	if t.Status == models.StatusOpen {
		h.tickets.SetStatus(c.Request.Context(), id, models.StatusInProgress)
	}

	c.JSON(http.StatusCreated, msg)
}

// ─── Admin: удалить / закрыть ─────────────────────────────────────────────────

func (h *Handler) AdminDeleteTicket(c *gin.Context) {
	id := parseID(c)
	if err := h.tickets.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) AdminCloseTicket(c *gin.Context) {
	id := parseID(c)
	adminLogin := getAdminName(c)
	if err := h.tickets.CloseByAdmin(c.Request.Context(), id, adminLogin, "admin"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "closed", "closed_by": adminLogin})
}

// getAdminName — достаёт логин администратора из JWT токена
func getAdminName(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if len(header) > 7 {
		claims, err := auth.ParseToken(header[7:])
		if err == nil && claims.Login != "" {
			return claims.Login
		}
	}
	return "Администратор"
}

func (h *Handler) GetIPLogs(c *gin.Context) {
	logs, err := h.admins.GetIPLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// ─── Export ───────────────────────────────────────────────────────────────────

func (h *Handler) ExportTickets(c *gin.Context) {
	format := c.Query("format")
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
	if format == "xml" {
		exportXML(c, tickets)
		return
	}
	c.JSON(http.StatusOK, gin.H{"format": format, "tickets": tickets})
}

type xmlTickets struct {
	XMLName xml.Name    `xml:"Tickets"`
	Items   []xmlTicket `xml:"Ticket"`
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
			ID: t.ID, FirstName: t.FirstName, LastName: t.LastName,
			Division: t.Division, Description: t.Description,
			Priority: string(t.Priority), Status: string(t.Status),
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
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
