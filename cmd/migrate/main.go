package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
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
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch {
	case *up:
		fmt.Println("Running all pending migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully!")

	case *down:
		fmt.Println("Rolling back last migration...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed!")

	case *reset:
		fmt.Println("Rolling back all migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("All migrations rolled back!")

	case *steps != 0:
		direction := "up"
		if *steps < 0 {
			direction = "down"
		}
		fmt.Printf("Running %d migration(s) %s...\n", abs(*steps), direction)
		if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Done!")

	case *force >= 0:
		fmt.Printf("Forcing version to %d...\n", *force)
		if err := m.Force(*force); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Println("Version forced!")

	case *version:
		v, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("📋 No migrations applied yet")
				return
			}
			log.Fatalf("Failed to get version: %v", err)
		}
		status := ""
		if dirty {
			status = " (dirty)"
		}
		fmt.Printf("📋 Current version: %d%s\n", v, status)

	default:
		fmt.Println("Formify Migration Tool")
		fmt.Println("======================")
		fmt.Println("Usage:")
		fmt.Println("  go run ./cmd/migrate -up        Apply all pending migrations")
		fmt.Println("  go run ./cmd/migrate -down      Rollback last migration")
		fmt.Println("  go run ./cmd/migrate -reset     Rollback all migrations")
		fmt.Println("  go run ./cmd/migrate -version   Show current migration version")
		fmt.Println("  go run ./cmd/migrate -steps N   Run N migrations (negative for down)")
		fmt.Println("  go run ./cmd/migrate -force N   Force set version to N")
		fmt.Println("")
		fmt.Println("Or use make commands:")
		fmt.Println("  make migrate-up      Apply all pending migrations")
		fmt.Println("  make migrate-down    Rollback last migration")
		fmt.Println("  make migrate-reset   Rollback all migrations")
		fmt.Println("  make migrate-status  Show current migration version")
		fmt.Println("  make migrate-create  Create a new migration file")
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
