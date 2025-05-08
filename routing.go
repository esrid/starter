package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"template/internal/handlers"
	"template/internal/repository"
	"template/internal/services"
)

type HandlerRegistery struct {
	UserHandler handlers.UserHandler
	Middleware  *handlers.Middleware
	logger      *slog.Logger
}

func NewHandlerRegistery(conn *sql.DB, logger *slog.Logger) *HandlerRegistery {
	sr := repository.SessionRepository{DB: conn}
	ss := services.NewSessionService(&sr)
	ur := repository.NewUserRepo(conn, logger)
	us := services.NewUserService(ur)
	uh := handlers.UserHandler{US: us, SS: ss}

	middleware := handlers.NewMiddleware(us, ss)
	return &HandlerRegistery{
		UserHandler: uh,
		Middleware:  middleware,
		logger:      logger,
	}
}

func (s *HandlerRegistery) Routes() http.Handler {
	root := http.NewServeMux()

	s.mountPublicRoutes(root)
	s.mountProtectedRoutes(root)

	return s.Middleware.Chain(root,
		s.Middleware.WithLogging(s.logger),
		s.Middleware.RateLimitMiddleware,
		s.Middleware.SecurityHeadersMiddleware,
	)
}

func (s *HandlerRegistery) mountPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "salut // home")
	})
	mux.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "salut from home")
	})
	mux.HandleFunc("/login", handlers.Home)
	mux.HandleFunc("/register", handlers.Home)
}

func (s *HandlerRegistery) mountProtectedRoutes(mux *http.ServeMux) {
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "out app")
	})
	protectedMux.HandleFunc("/dashboard", handlers.Home)
	protectedMux.HandleFunc("/account", handlers.Home)

	handler := s.Middleware.Chain(protectedMux,
		s.Middleware.AuthMiddleware,
		s.Middleware.CSRFMiddleware,
		s.Middleware.SessionRefreshMiddleware,
	)

	mux.Handle("/app/", http.StripPrefix("/app", handler))
}

func startServer(s *http.Server, serverErr chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		serverErr <- err
	}
}

// gracefulShutdown handles server shutdown with a timeout
func gracefulShutdown(ctx context.Context, s *http.Server) error {
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeout); err != nil {
		return fmt.Errorf("server failed to shutdown: %w", err)
	}
	return nil
}
