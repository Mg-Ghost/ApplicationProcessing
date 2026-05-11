package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"meddoc/internal/models"
)

type TicketRepo struct{ db *pgxpool.Pool }

func NewTicketRepo(db *pgxpool.Pool) *TicketRepo { return &TicketRepo{db} }

func (r *TicketRepo) Create(ctx context.Context, t *models.Ticket) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO tickets
		 (user_id, first_name, last_name, phone, position, room, division,
		  description, inventory_number, ip_address, priority, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING id, created_at, updated_at`,
		t.UserID, t.FirstName, t.LastName, t.Phone, t.Position, t.Room,
		t.Division, t.Description, t.InventoryNumber, t.IPAddress, t.Priority, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TicketRepo) GetByID(ctx context.Context, id int64) (*models.Ticket, error) {
	t := &models.Ticket{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, first_name, last_name, phone, position, room, division,
		 description, inventory_number, ip_address, priority, status,
		 admin_comment, auto_escalated, created_at, updated_at,
		 closed_by_admin, closed_by_role
		 FROM tickets WHERE id=$1`, id,
	).Scan(
		&t.ID, &t.UserID, &t.FirstName, &t.LastName, &t.Phone, &t.Position,
		&t.Room, &t.Division, &t.Description, &t.InventoryNumber, &t.IPAddress,
		&t.Priority, &t.Status, &t.AdminComment, &t.AutoEscalated,
		&t.CreatedAt, &t.UpdatedAt, &t.ClosedByAdmin, &t.ClosedByRole,
	)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}
	return t, nil
}

func (r *TicketRepo) ListByUser(ctx context.Context, userID int64) ([]*models.Ticket, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, first_name, last_name, phone, position, room, division,
		 description, inventory_number, ip_address, priority, status,
		 admin_comment, auto_escalated, created_at, updated_at,
		 closed_by_admin, closed_by_role
		 FROM tickets WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *TicketRepo) ListAll(ctx context.Context, f models.TicketFilter) ([]*models.Ticket, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.Division != "" {
		where = append(where, fmt.Sprintf("division=$%d", i)); args = append(args, f.Division); i++
	}
	if f.Priority != "" {
		where = append(where, fmt.Sprintf("priority=$%d", i)); args = append(args, f.Priority); i++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", i)); args = append(args, f.Status); i++
	}
	if f.DateFrom != "" {
		where = append(where, fmt.Sprintf("created_at >= $%d", i)); args = append(args, f.DateFrom); i++
	}
	if f.DateTo != "" {
		where = append(where, fmt.Sprintf("created_at <= $%d", i)); args = append(args, f.DateTo+" 23:59:59"); i++
	}

	sortCol := "created_at"
	if map[string]bool{"created_at": true, "priority": true, "division": true}[f.SortBy] {
		sortCol = f.SortBy
	}
	sortDir := "DESC"
	if strings.ToUpper(f.SortOrder) == "ASC" {
		sortDir = "ASC"
	}

	q := fmt.Sprintf(
		`SELECT id, user_id, first_name, last_name, phone, position, room, division,
		 description, inventory_number, ip_address, priority, status,
		 admin_comment, auto_escalated, created_at, updated_at,
		 closed_by_admin, closed_by_role
		 FROM tickets WHERE %s ORDER BY %s %s`,
		strings.Join(where, " AND "), sortCol, sortDir,
	)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *TicketRepo) Update(ctx context.Context, id int64, req *models.UpdateTicketRequest) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tickets SET
		 phone            = COALESCE(NULLIF($1,''), phone),
		 position         = COALESCE(NULLIF($2,''), position),
		 room             = COALESCE(NULLIF($3,''), room),
		 description      = COALESCE(NULLIF($4,''), description),
		 inventory_number = COALESCE(NULLIF($5,''), inventory_number),
		 ip_address       = COALESCE(NULLIF($6,''), ip_address),
		 priority         = COALESCE(NULLIF($7,''), priority),
		 updated_at       = NOW()
		 WHERE id=$8`,
		req.Phone, req.Position, req.Room, req.Description,
		req.InventoryNumber, req.IPAddress, string(req.Priority), id,
	)
	return err
}

func (r *TicketRepo) SetStatus(ctx context.Context, id int64, status models.TicketStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tickets SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

// CloseByAdmin — закрывает заявление, записывает кто закрыл и роль
func (r *TicketRepo) CloseByAdmin(ctx context.Context, id int64, closedBy string, role string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tickets SET status='closed', closed_by_admin=$1, closed_by_role=$2, updated_at=NOW() WHERE id=$3`,
		closedBy, role, id)
	return err
}

func (r *TicketRepo) AddComment(ctx context.Context, id int64, comment string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tickets SET admin_comment=$1, status='in_progress', updated_at=NOW() WHERE id=$2`,
		comment, id)
	return err
}

func (r *TicketRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tickets WHERE id=$1`, id)
	return err
}

func (r *TicketRepo) EscalateOldTickets(ctx context.Context) (int64, error) {
	res, err := r.db.Exec(ctx,
		`UPDATE tickets SET priority='high', auto_escalated=TRUE, updated_at=NOW()
		 WHERE status IN ('open','in_progress')
		   AND priority != 'high'
		   AND auto_escalated = FALSE
		   AND created_at <= NOW() - INTERVAL '7 days'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func scanTickets(rows interface{ Next() bool; Scan(...any) error }) ([]*models.Ticket, error) {
	var list []*models.Ticket
	for rows.Next() {
		t := &models.Ticket{}
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.FirstName, &t.LastName, &t.Phone, &t.Position,
			&t.Room, &t.Division, &t.Description, &t.InventoryNumber, &t.IPAddress,
			&t.Priority, &t.Status, &t.AdminComment, &t.AutoEscalated,
			&t.CreatedAt, &t.UpdatedAt, &t.ClosedByAdmin, &t.ClosedByRole,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}