package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "meddoc_user"),
		getenv("DB_PASSWORD", "supersecret"),
		getenv("DB_NAME", "meddoc_db"),
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return pool, nil
}

func RunMigrations(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), schema)
	return err
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    division      TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin_users (
    id            BIGSERIAL PRIMARY KEY,
    login         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tickets (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT      NOT NULL REFERENCES users(id),
    first_name       TEXT        NOT NULL,
    last_name        TEXT        NOT NULL,
    phone            TEXT        NOT NULL,
    position         TEXT        NOT NULL,
    room             TEXT        NOT NULL,
    division         TEXT        NOT NULL,
    description      TEXT        NOT NULL,
    inventory_number TEXT,
    ip_address       TEXT,
    priority         TEXT        NOT NULL DEFAULT 'medium',
    status           TEXT        NOT NULL DEFAULT 'open',
    admin_comment    TEXT        NOT NULL DEFAULT '',
    auto_escalated   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ip_logs (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    login      TEXT        NOT NULL,
    ip_address TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_user_id  ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status   ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_division ON tickets(division);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id          BIGSERIAL PRIMARY KEY,
    ticket_id   BIGINT      NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author      TEXT        NOT NULL,
    author_name TEXT        NOT NULL,
    text        TEXT        NOT NULL,
    read_by_user  BOOLEAN NOT NULL DEFAULT FALSE,
    read_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket_id ON ticket_messages(ticket_id);

-- Добавляем колонки если таблица уже существует (для существующих БД)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='ticket_messages' AND column_name='read_by_user') THEN
    ALTER TABLE ticket_messages ADD COLUMN read_by_user BOOLEAN NOT NULL DEFAULT FALSE;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='ticket_messages' AND column_name='read_by_admin') THEN
    ALTER TABLE ticket_messages ADD COLUMN read_by_admin BOOLEAN NOT NULL DEFAULT FALSE;
  END IF;
END $$;
`
