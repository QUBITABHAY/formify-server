package main

import (
	"flag"
	"log"
	"os"

	"formify/server/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	if err := logger.InitFromEnv(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	log := logger.GetLogger()
	sugar := logger.GetSugaredLogger()

	godotenv.Load()

	up := flag.Bool("up", false, "Run all pending migrations")
	down := flag.Bool("down", false, "Rollback the last migration")
	reset := flag.Bool("reset", false, "Rollback all migrations")
	version := flag.Bool("version", false, "Show current migration version")
	steps := flag.Int("steps", 0, "Number of migrations to run (positive=up, negative=down)")
	force := flag.Int("force", -1, "Force set version (use with caution)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	migrationsPath := "file://internal/database/migrations"

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		log.Fatal("Failed to create migrate instance", logger.ToField("error", err))
	}
	defer m.Close()

	switch {
	case *up:
		sugar.Info("Running all pending migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration failed", logger.ToField("error", err))
		}
		sugar.Info("Migrations applied successfully!")

	case *down:
		sugar.Info("Rolling back last migration...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Rollback failed", logger.ToField("error", err))
		}
		sugar.Info("Rollback completed!")

	case *reset:
		sugar.Info("Rolling back all migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Reset failed", logger.ToField("error", err))
		}
		sugar.Info("All migrations rolled back!")

	case *steps != 0:
		direction := "up"
		if *steps < 0 {
			direction = "down"
		}
		sugar.Infof("Running %d migration(s) %s...", abs(*steps), direction)
		if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration failed", logger.ToField("error", err))
		}
		sugar.Info("Done!")

	case *force >= 0:
		sugar.Infof("Forcing version to %d...", *force)
		if err := m.Force(*force); err != nil {
			log.Fatal("Force failed", logger.ToField("error", err))
		}
		sugar.Info("Version forced!")

	case *version:
		v, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				sugar.Info("No migrations applied yet")
				return
			}
			log.Fatal("Failed to get version", logger.ToField("error", err))
		}
		status := ""
		if dirty {
			status = " (dirty)"
		}
		sugar.Infof("Current version: %d%s", v, status)

	default:
		sugar.Info("Formify Migration Tool")
		sugar.Info("======================")
		sugar.Info("Usage:")
		sugar.Info("  go run ./cmd/migrate -up        Apply all pending migrations")
		sugar.Info("  go run ./cmd/migrate -down      Rollback last migration")
		sugar.Info("  go run ./cmd/migrate -reset     Rollback all migrations")
		sugar.Info("  go run ./cmd/migrate -version   Show current migration version")
		sugar.Info("  go run ./cmd/migrate -steps N   Run N migrations (negative for down)")
		sugar.Info("  go run ./cmd/migrate -force N   Force set version to N")
		sugar.Info("")
		sugar.Info("Or use make commands:")
		sugar.Info("  make migrate-up      Apply all pending migrations")
		sugar.Info("  make migrate-down    Rollback last migration")
		sugar.Info("  make migrate-reset   Rollback all migrations")
		sugar.Info("  make migrate-status  Show current migration version")
		sugar.Info("  make migrate-create  Create a new migration file")
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
