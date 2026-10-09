// Command server runs the DnD API and websocket gateway.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"dndproject/backend/internal/config"
	"dndproject/backend/internal/database"
	"dndproject/backend/internal/handler"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/ws"
)

func main() {
	// Repo root first, then the current directory; whichever exists wins.
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	cfg := config.Load()

	db, err := database.Open(cfg.DBDSN)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	go purgeExpiredSessions(ctx, repository.NewSessionRepository(db))

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler.NewRouter(cfg, db, ws.NewHub()),
		// No WriteTimeout: it would cut long-lived websocket connections.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}
	slog.Info("server stopped")
}

// purgeExpiredSessions keeps the sessions table from growing without bound.
// Expired rows are already ignored by the lookup query, so the interval is a
// housekeeping choice, not a security one.
func purgeExpiredSessions(ctx context.Context, sessions *repository.SessionRepository) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			removed, err := sessions.DeleteExpired(ctx)
			if err != nil {
				slog.Error("purge sessions", "error", err)
				continue
			}
			if removed > 0 {
				slog.Info("purged expired sessions", "count", removed)
			}
		}
	}
}
