package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"time"

	"template/internal/repository"
)

var (
	ErrSessionExpired = errors.New("session has expired")
	ErrInvalidSession = errors.New("invalid session")
)

type SessionService struct {
	repo *repository.SessionRepository
	// Session configuration
	sessionDuration time.Duration
	maxSessions     int
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{
		repo:            repo,
		sessionDuration: 24 * time.Hour, // Default session duration
		maxSessions:     5,              // Maximum concurrent sessions per user
	}
}

// GenerateSessionToken generates a cryptographically secure random token
func (s *SessionService) GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session for a user with security best practices
func (s *SessionService) CreateSession(ctx context.Context, userID string, ipAddress net.IP, userAgent string) (string, error) {
	// Generate secure session token
	cookieHash, err := s.GenerateSessionToken()
	if err != nil {
		return "", err
	}

	session := repository.Session{
		UserID:     userID,
		CookieHash: cookieHash,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  &time.Time{},
	}

	// Set expiration time
	expiresAt := time.Now().Add(s.sessionDuration)
	session.ExpiresAt = &expiresAt

	// Create the session
	cookieHash, err = s.repo.Create(ctx, session)
	if err != nil {
		return "", err
	}

	return cookieHash, nil
}

// ValidateSession validates a session and checks for expiration
func (s *SessionService) ValidateSession(ctx context.Context, cookieHash string) (repository.Session, error) {
	session, err := s.repo.GetByCookieHash(ctx, cookieHash)
	if err != nil {
		return repository.Session{}, ErrInvalidSession
	}

	// Check if session has expired
	if session.ExpiresAt != nil && time.Now().After(*session.ExpiresAt) {
		// Automatically delete expired session
		_ = s.repo.DeleteByCookieHash(ctx, cookieHash)
		return repository.Session{}, ErrSessionExpired
	}

	return session, nil
}

// RefreshSession extends the session duration
func (s *SessionService) RefreshSession(ctx context.Context, cookieHash string) error {
	// Validate session first
	_, err := s.ValidateSession(ctx, cookieHash)
	if err != nil {
		return err
	}

	// Set new expiration time
	newExpiryTime := time.Now().Add(s.sessionDuration)
	return s.repo.UpdateExpiry(ctx, cookieHash, newExpiryTime)
}

// RevokeSession invalidates a session
func (s *SessionService) RevokeSession(ctx context.Context, cookieHash string) error {
	return s.repo.DeleteByCookieHash(ctx, cookieHash)
}

// RevokeAllUserSessions invalidates all sessions for a given user
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.repo.DeleteByUserID(ctx, userID)
}
