package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"meddoc/internal/models"
)

type MessageRepo struct{ db *pgxpool.Pool }

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo { return &MessageRepo{db} }

func (r *MessageRepo) Add(ctx context.Context, m *models.TicketMessage) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO ticket_messages (ticket_id, author, author_name, text, read_by_user, read_by_admin)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		m.TicketID, m.Author, m.AuthorName, m.Text, m.ReadByUser, m.ReadByAdmin,
	).Scan(&m.ID, &m.CreatedAt)
}

func (r *MessageRepo) ListByTicket(ctx context.Context, ticketID int64) ([]*models.TicketMessage, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticket_id, author, author_name, text, read_by_user, read_by_admin, created_at
		 FROM ticket_messages
		 WHERE ticket_id = $1
		 ORDER BY created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*models.TicketMessage
	for rows.Next() {
		m := &models.TicketMessage{}
		if err := rows.Scan(&m.ID, &m.TicketID, &m.Author, &m.AuthorName, &m.Text,
			&m.ReadByUser, &m.ReadByAdmin, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// MarkReadByUser — отмечает все сообщения от админа как прочитанные пользователем
func (r *MessageRepo) MarkReadByUser(ctx context.Context, ticketID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE ticket_messages SET read_by_user = TRUE
		 WHERE ticket_id = $1 AND author = 'admin' AND read_by_user = FALSE`,
		ticketID,
	)
	return err
}

// MarkReadByAdmin — отмечает все сообщения от пользователя как прочитанные админом
func (r *MessageRepo) MarkReadByAdmin(ctx context.Context, ticketID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE ticket_messages SET read_by_admin = TRUE
		 WHERE ticket_id = $1 AND author = 'user' AND read_by_admin = FALSE`,
		ticketID,
	)
	return err
}

// UnreadCountForUser — кол-во непрочитанных сообщений от админа для одной заявки
func (r *MessageRepo) UnreadCountForUser(ctx context.Context, ticketID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM ticket_messages
		 WHERE ticket_id = $1 AND author = 'admin' AND read_by_user = FALSE`,
		ticketID,
	).Scan(&n)
	return n, err
}

// UnreadCountForAdmin — кол-во непрочитанных от пользователей для одной заявки
func (r *MessageRepo) UnreadCountForAdmin(ctx context.Context, ticketID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM ticket_messages
		 WHERE ticket_id = $1 AND author = 'user' AND read_by_admin = FALSE`,
		ticketID,
	).Scan(&n)
	return n, err
}

// UnreadPerTicketForAdmin — возвращает map[ticketID]count для всех заявок
func (r *MessageRepo) UnreadPerTicketForAdmin(ctx context.Context) (map[int64]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT ticket_id, COUNT(*) FROM ticket_messages
		 WHERE author = 'user' AND read_by_admin = FALSE
		 GROUP BY ticket_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64]int{}
	for rows.Next() {
		var id int64
		var n int
		rows.Scan(&id, &n)
		m[id] = n
	}
	return m, nil
}

// UnreadPerTicketForUser — возвращает map[ticketID]count для пользователя
func (r *MessageRepo) UnreadPerTicketForUser(ctx context.Context, userID int64) (map[int64]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tm.ticket_id, COUNT(*) FROM ticket_messages tm
		 JOIN tickets t ON t.id = tm.ticket_id
		 WHERE t.user_id = $1 AND tm.author = 'admin' AND tm.read_by_user = FALSE
		 GROUP BY tm.ticket_id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64]int{}
	for rows.Next() {
		var id int64
		var n int
		rows.Scan(&id, &n)
		m[id] = n
	}
	return m, nil
}
