package models

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
	StatusCancelled  TicketStatus = "cancelled"
)

// ─── User ─────────────────────────────────────────────────────────────────────

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Division     string    `json:"division"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required,min=2"`
	LastName  string `json:"last_name"  binding:"required,min=2"`
	Division  string `json:"division"   binding:"required"`
	Password  string `json:"password"   binding:"required,min=6"`
}

type LoginRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	Password  string `json:"password"   binding:"required"`
}

type AdminLoginRequest struct {
	Login     string `json:"login"      binding:"required"`
	Password  string `json:"password"   binding:"required,min=6"`
	SecretKey string `json:"secret_key" binding:"required,min=1"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Division  string `json:"division"`
	Password  string `json:"password"`
}

// ─── Ticket Message (переписка внутри заявки) ─────────────────────────────────

type MessageAuthor string

const (
	AuthorUser  MessageAuthor = "user"
	AuthorAdmin MessageAuthor = "admin"
)

type TicketMessage struct {
	ID           int64         `json:"id"`
	TicketID     int64         `json:"ticket_id"`
	Author       MessageAuthor `json:"author"`
	AuthorName   string        `json:"author_name"`
	Text         string        `json:"text"`
	ReadByUser   bool          `json:"read_by_user"`
	ReadByAdmin  bool          `json:"read_by_admin"`
	CreatedAt    time.Time     `json:"created_at"`
}

// ─── Ticket ───────────────────────────────────────────────────────────────────

type Ticket struct {
	ID              int64            `json:"id"`
	UserID          int64            `json:"user_id"`
	FirstName       string           `json:"first_name"`
	LastName        string           `json:"last_name"`
	Phone           string           `json:"phone"`
	Position        string           `json:"position"`
	Room            string           `json:"room"`
	Division        string           `json:"division"`
	Description     string           `json:"description"`
	InventoryNumber string           `json:"inventory_number"`
	IPAddress       string           `json:"ip_address"`
	Priority        Priority         `json:"priority"`
	Status          TicketStatus     `json:"status"`
	AdminComment    string           `json:"admin_comment"`
	ClosedByAdmin   string           `json:"closed_by_admin"`   // имя/логин кто закрыл
	ClosedByRole    string           `json:"closed_by_role"`    // "admin" или "user"
	AutoEscalated   bool             `json:"auto_escalated"`
	Messages        []*TicketMessage `json:"messages,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type CreateTicketRequest struct {
	FirstName       string `json:"first_name"       binding:"required"`
	LastName        string `json:"last_name"        binding:"required"`
	Phone           string `json:"phone"            binding:"required"`
	Position        string `json:"position"         binding:"required"`
	Room            string `json:"room"             binding:"required"`
	Division        string `json:"division"         binding:"required"`
	Description     string `json:"description"      binding:"required,min=10"`
	InventoryNumber string `json:"inventory_number"`
	IPAddress       string `json:"ip_address"`
	// Приоритет задаёт только администратор, пользователь всегда получает Medium
}

type UpdateTicketRequest struct {
	Phone           string `json:"phone"`
	Position        string `json:"position"`
	Room            string `json:"room"`
	Description     string `json:"description"`
	InventoryNumber string `json:"inventory_number"`
	IPAddress       string `json:"ip_address"`
	// Приоритет убран — только администратор может менять
}

// AdminUpdatePriorityRequest — только для администратора
type AdminUpdatePriorityRequest struct {
	Priority Priority `json:"priority" binding:"required,oneof=low medium high"`
}

type AddCommentRequest struct {
	Comment string `json:"comment" binding:"required,min=1"`
}

type AddUserReplyRequest struct {
	Text string `json:"text" binding:"required,min=1"`
}

type TicketFilter struct {
	Division  string
	Priority  string
	Status    string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
}

// ─── IP Log ───────────────────────────────────────────────────────────────────

type IPLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Login     string    `json:"login"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
