package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"meddoc/internal/models"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (first_name, last_name, division, password_hash, role)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		u.FirstName, u.LastName, u.Division, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, first_name, last_name, division, password_hash, role, created_at
		 FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Division, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByName(ctx context.Context, firstName string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, first_name, last_name, division, password_hash, role, created_at
		 FROM users WHERE first_name=$1 LIMIT 1`, firstName,
	).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Division, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, id int64, req *models.UpdateProfileRequest) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET
		 first_name    = COALESCE(NULLIF($1,''), first_name),
		 last_name     = COALESCE(NULLIF($2,''), last_name),
		 division      = COALESCE(NULLIF($3,''), division),
		 password_hash = COALESCE(NULLIF($4,''), password_hash)
		 WHERE id=$5`,
		req.FirstName, req.LastName, req.Division, req.Password, id,
	)
	return err
}

// ─── Admin ───────────────────────────────────────────────────────────────────

type AdminRepo struct{ db *pgxpool.Pool }

func NewAdminRepo(db *pgxpool.Pool) *AdminRepo { return &AdminRepo{db} }

func (r *AdminRepo) GetByLogin(ctx context.Context, login string) (int64, string, error) {
	var id int64
	var hash string
	err := r.db.QueryRow(ctx,
		`SELECT id, password_hash FROM admin_users WHERE login=$1`, login,
	).Scan(&id, &hash)
	return id, hash, err
}

func (r *AdminRepo) LogIP(ctx context.Context, log *models.IPLog) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO ip_logs (user_id, login, ip_address) VALUES ($1,$2,$3)`,
		log.UserID, log.Login, log.IPAddress,
	)
	return err
}

func (r *AdminRepo) GetIPLogs(ctx context.Context) ([]*models.IPLog, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, login, ip_address, created_at FROM ip_logs ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*models.IPLog
	for rows.Next() {
		l := &models.IPLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Login, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
