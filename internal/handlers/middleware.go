package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"template/internal/services"
	"template/utils"
)

type userctx string

const userkey userctx = "user"

type Middleware struct {
	userService    *services.UserService
	sessionService *services.SessionService
}

func NewMiddleware(us *services.UserService, ss *services.SessionService) *Middleware {
	return &Middleware{
		userService:    us,
		sessionService: ss,
	}
}

func (m *Middleware) Chain(h http.Handler, mw ...Mw) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

type Mw func(http.Handler) http.Handler

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *Middleware) WithLogging(logger *slog.Logger) Mw {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			w.Header().Set("X-Request-ID", requestID)

			logger.Info("request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
			)

			next.ServeHTTP(rw, r)

			logger.Info("request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", http.StatusText(rw.statusCode),
				"duration", time.Since(start),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := m.sessionService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				MaxAge:   -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := m.userService.UR.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			_ = m.sessionService.RevokeSession(r.Context(), cookie.Value)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userkey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		csrfToken := r.FormValue("csrf_token")
		if csrfToken == "" || csrfToken != csrfCookie.Value {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RateLimitMiddleware(next http.Handler) http.Handler {
	type client struct {
		count     int
		lastReset time.Time
	}
	clients := make(map[string]*client)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetIPAddressBytes(r)
		if ip == nil {
			http.Error(w, "Could not determine client IP", http.StatusBadRequest)
			return
		}

		ipStr := string(ip)
		now := time.Now()

		c, exists := clients[ipStr]
		if !exists {
			c = &client{lastReset: now}
			clients[ipStr] = c
		}

		if now.Sub(c.lastReset) > time.Minute {
			c.count = 0
			c.lastReset = now
		}

		if c.count >= 100 {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		c.count++
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Pragma", "no-cache")

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) SessionRefreshMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil {
			err = m.sessionService.RefreshSession(r.Context(), cookie.Value)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    cookie.Value,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
					MaxAge:   int(24 * time.Hour.Seconds()),
				})
			}
		}

		next.ServeHTTP(w, r)
	})
}
