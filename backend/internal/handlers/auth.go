package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"meddoc/internal/auth"
	"meddoc/internal/middleware"
	"meddoc/internal/models"
	"meddoc/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	users   *repository.UserRepo
	admins  *repository.AdminRepo
	tickets *repository.TicketRepo
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{
		users:   repository.NewUserRepo(db),
		admins:  repository.NewAdminRepo(db),
		tickets: repository.NewTicketRepo(db),
	}
}

// ─── Register ────────────────────────────────────────────────────────────────

func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}

	u := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Division:     req.Division,
		PasswordHash: string(hash),
		Role:         models.RoleUser,
	}
	if err := h.users.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	token, _ := auth.GenerateToken(u.ID, string(u.Role))
	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: u})
}

// ─── Login ───────────────────────────────────────────────────────────────────

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.users.GetByName(c.Request.Context(), req.FirstName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := auth.GenerateToken(u.ID, string(u.Role))
	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: u})
}

// ─── Admin Login ─────────────────────────────────────────────────────────────

func (h *Handler) AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate secret key
	if req.SecretKey != os.Getenv("ADMIN_SECRET_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid secret key"})
		return
	}

	id, hash, err := h.admins.GetByLogin(c.Request.Context(), req.Login)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Log IP
	ip := c.ClientIP()
	h.admins.LogIP(c.Request.Context(), &models.IPLog{
		UserID:    id,
		Login:     req.Login,
		IPAddress: ip,
	})

	token, _ := auth.GenerateToken(id, "admin")
	c.JSON(http.StatusOK, gin.H{"token": token, "role": "admin", "ip": ip})
}

// ─── Profile ─────────────────────────────────────────────────────────────────

func (h *Handler) GetProfile(c *gin.Context) {
	id := middleware.GetUserID(c)
	u, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
			return
		}
		req.Password = string(hash)
	}

	id := middleware.GetUserID(c)
	if err := h.users.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
