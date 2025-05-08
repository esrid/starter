package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash sql.NullString
	GoogleID     sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type UserRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepo(db *sql.DB, log *slog.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: log,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, u *User) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        INSERT INTO users (email, password_hash, google_id, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, email, password_hash, google_id, created_at, updated_at`,
		u.Email, u.PasswordHash, u.GoogleID)
	user := &User{}
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.GoogleID, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, google_id, created_at, updated_at
        FROM users WHERE id = $1`, id)
	u := &User{}
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, google_id, created_at, updated_at
        FROM users WHERE email = $1`, email)
	u := &User{}
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetUserByGoogleID(ctx context.Context, gid string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, google_id, created_at, updated_at
        FROM users WHERE google_id = $1`, sql.NullString{String: gid, Valid: true})
	u := &User{}
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetAllUsers(ctx context.Context) ([]*User, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, email, password_hash, google_id, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, u *User) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE users SET email = $1, password_hash = $2, google_id = $3, updated_at = NOW() WHERE id = $4`,
		u.Email, u.PasswordHash, u.GoogleID, u.ID)
	return err
}

func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepo) GetUserBySessionID(ctx context.Context, sid string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT u.id, u.email, u.password_hash, u.google_id, u.created_at, u.updated_at
        FROM users u INNER JOIN sessions s ON u.id = s.user_id WHERE s.session_id = $1`, sid)
	u := &User{}
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
