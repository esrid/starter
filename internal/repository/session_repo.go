package repository

import (
	"context"
	"database/sql"
	"net"
	"time"
)

type Session struct {
	ID         string
	UserID     string
	CookieHash string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	IPAddress  net.IP
	UserAgent  string
}

type SessionRepository struct {
	DB *sql.DB
}

func (ss *SessionRepository) Create(ctx context.Context, s Session) (string, error) {
	var cookieHash string
	err := ss.DB.QueryRowContext(ctx, `
        INSERT INTO sessions (user_id, cookie_hash, created_at, expires_at, ip_address, user_agent)
        VALUES ($1, $2, NOW(), NULL, $3, $4)
        ON CONFLICT (user_id) DO UPDATE SET
            cookie_hash = EXCLUDED.cookie_hash,
            created_at = NOW(),
            expires_at = NULL,
            ip_address = EXCLUDED.ip_address,
            user_agent = EXCLUDED.user_agent
        RETURNING cookie_hash
    `, s.UserID, s.CookieHash, s.IPAddress, s.UserAgent).Scan(&cookieHash)
	if err != nil {
		return "", err
	}
	return cookieHash, nil
}

func (ss *SessionRepository) GetByCookieHash(ctx context.Context, cookieHash string) (Session, error) {
	var s Session
	err := ss.DB.QueryRowContext(ctx, `
        SELECT id, user_id, cookie_hash, created_at, expires_at, ip_address, user_agent
        FROM sessions
        WHERE cookie_hash = $1
    `, cookieHash).Scan(&s.ID, &s.UserID, &s.CookieHash, &s.CreatedAt, &s.ExpiresAt, &s.IPAddress, &s.UserAgent)
	if err != nil {
		return Session{}, err
	}
	return s, nil
}

func (ss *SessionRepository) DeleteByCookieHash(ctx context.Context, cookieHash string) error {
	_, err := ss.DB.ExecContext(ctx, `
        DELETE FROM sessions
        WHERE cookie_hash = $1
    `, cookieHash)
	return err
}

func (ss *SessionRepository) UpdateExpiry(ctx context.Context, cookieHash string, expiresAt time.Time) error {
	_, err := ss.DB.ExecContext(ctx, `
        UPDATE sessions
        SET expires_at = $1
        WHERE cookie_hash = $2
    `, expiresAt, cookieHash)
	return err
}
func (ss *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := ss.DB.ExecContext(ctx, `
        DELETE FROM sessions
        WHERE user_id = $1
    `, userID)
	return err
}
