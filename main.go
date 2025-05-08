package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"template/config"
	"template/db"
)

func main() {
	cfg := config.Load()
	logger := config.NewSlog(cfg.Env)

	conn, err := db.NewDB(cfg.Database.String())
	if err != nil {
		logger.Error("db connection failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	hr := NewHandlerRegistery(conn, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	serverErr := make(chan error, 1)
	var wg sync.WaitGroup
	defer wg.Wait()

	s := &http.Server{
		Addr:     ":80",
		Handler:  hr.Routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	wg.Add(1)
	go startServer(s, serverErr, &wg)

	select {
	case err := <-serverErr:
		logger.Error("server failed", slog.String("error", err.Error()))
	case <-ctx.Done():
	}

	if err := gracefulShutdown(ctx, s); err != nil {
		logger.Error("server failed to shutdown", slog.String("error", err.Error()))
	}
}
