// Command migrate applies or rolls back the embedded MySQL migrations.
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate down 1
//	go run ./cmd/migrate status
package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"dndproject/backend/internal/config"
	"dndproject/backend/internal/database"
	"dndproject/backend/migrations"
)

func main() {
	// Repo root first, then the current directory; whichever exists wins.
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	cfg := config.Load()

	db, err := database.Open(cfg.DBDSN)
	if err != nil {
		fail("connect database", err)
	}
	defer db.Close()

	ctx := context.Background()
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		err = database.MigrateUp(ctx, db, migrations.FS)
	case "down":
		steps := 1
		if len(os.Args) > 2 {
			if n, convErr := strconv.Atoi(os.Args[2]); convErr == nil && n > 0 {
				steps = n
			}
		}
		err = database.MigrateDown(ctx, db, migrations.FS, steps)
	case "status":
		current, pending, statusErr := database.MigrationStatus(ctx, db, migrations.FS)
		if statusErr != nil {
			fail("status", statusErr)
		}
		slog.Info("migration status", "current_version", current, "pending", pending)
		return
	default:
		slog.Error("unknown command", "command", cmd, "want", "up | down [steps] | status")
		os.Exit(2)
	}

	if err != nil {
		fail("migrate "+cmd, err)
	}
}

func fail(what string, err error) {
	slog.Error(what, "error", err)
	os.Exit(1)
}
