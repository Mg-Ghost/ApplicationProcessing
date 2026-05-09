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
		`INSERT INTO ticket_messages (ticket_id, author, author_name, text)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		m.TicketID, m.Author, m.AuthorName, m.Text,
	).Scan(&m.ID, &m.CreatedAt)
}

func (r *MessageRepo) ListByTicket(ctx context.Context, ticketID int64) ([]*models.TicketMessage, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticket_id, author, author_name, text, created_at
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
		if err := rows.Scan(&m.ID, &m.TicketID, &m.Author, &m.AuthorName, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}
