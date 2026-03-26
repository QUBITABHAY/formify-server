package main

import (
	"errors"
	"flag"
	stdlog "log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"formify/server/internal/logger"
)

type migrateOptions struct {
	up      bool
	down    bool
	reset   bool
	version bool
	steps   int
	force   int
}

func main() {
	if err := logger.InitFromEnv(); err != nil {
		stdlog.Fatalf("Failed to initialize logger: %v", err)
	}
	zapLog := logger.GetLogger()
	sugar := logger.GetSugaredLogger()

	defer func() {
		if err := logger.Close(); err != nil {
			zapLog.Error("Failed to close logger", logger.ToField("error", err))
		}
	}()

	if err := godotenv.Load(); err != nil {
		sugar.Warnf(".env file not loaded: %v", err)
	}

	up := flag.Bool("up", false, "Run all pending migrations")
	down := flag.Bool("down", false, "Rollback the last migration")
	reset := flag.Bool("reset", false, "Rollback all migrations")
	version := flag.Bool("version", false, "Show current migration version")
	steps := flag.Int("steps", 0, "Number of migrations to run (positive=up, negative=down)")
	force := flag.Int("force", -1, "Force set version (use with caution)")
	flag.Parse()

	opts := parseFlags(up, down, reset, version, steps, force)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		zapLog.Fatal("DATABASE_URL environment variable is required")
	}

	migrationsPath := "file://internal/database/migrations"

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		zapLog.Fatal("Failed to create migrate instance", logger.ToField("error", err))
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			zapLog.Error("Failed to close migration source", logger.ToField("error", sourceErr))
		}
		if dbErr != nil {
			zapLog.Error("Failed to close migration database", logger.ToField("error", dbErr))
		}
	}()

	runMigrations(m, opts, zapLog, sugar)
}

func parseFlags(up, down, reset, version *bool, steps, force *int) migrateOptions {
	return migrateOptions{
		up:      *up,
		down:    *down,
		reset:   *reset,
		version: *version,
		steps:   *steps,
		force:   *force,
	}
}

func runMigrations(m *migrate.Migrate, opts migrateOptions, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	switch {
	case opts.up:
		runUp(m, zapLog, sugar)
	case opts.down:
		runDown(m, zapLog, sugar)
	case opts.reset:
		runReset(m, zapLog, sugar)
	case opts.steps != 0:
		runSteps(m, opts.steps, zapLog, sugar)
	case opts.force >= 0:
		runForce(m, opts.force, zapLog, sugar)
	case opts.version:
		runVersion(m, zapLog, sugar)
	default:
		printUsage(sugar)
	}
}

func runUp(m *migrate.Migrate, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	sugar.Info("Running all pending migrations...")
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zapLog.Fatal("Migration failed", logger.ToField("error", err))
	}
	sugar.Info("Migrations applied successfully!")
}

func runDown(m *migrate.Migrate, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	sugar.Info("Rolling back last migration...")
	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zapLog.Fatal("Rollback failed", logger.ToField("error", err))
	}
	sugar.Info("Rollback completed!")
}

func runReset(m *migrate.Migrate, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	sugar.Info("Rolling back all migrations...")
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zapLog.Fatal("Reset failed", logger.ToField("error", err))
	}
	sugar.Info("All migrations rolled back!")
}

func runSteps(m *migrate.Migrate, steps int, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	direction := "up"
	if steps < 0 {
		direction = "down"
	}
	sugar.Infof("Running %d migration(s) %s...", abs(steps), direction)
	if err := m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zapLog.Fatal("Migration failed", logger.ToField("error", err))
	}
	sugar.Info("Done!")
}

func runForce(m *migrate.Migrate, force int, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	sugar.Infof("Forcing version to %d...", force)
	if err := m.Force(force); err != nil {
		zapLog.Fatal("Force failed", logger.ToField("error", err))
	}
	sugar.Info("Version forced!")
}

func runVersion(m *migrate.Migrate, zapLog *zap.Logger, sugar *zap.SugaredLogger) {
	v, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			sugar.Info("No migrations applied yet")
			return
		}
		zapLog.Fatal("Failed to get version", logger.ToField("error", err))
	}

	status := ""
	if dirty {
		status = " (dirty)"
	}
	sugar.Infof("Current version: %d%s", v, status)
}

func printUsage(sugar *zap.SugaredLogger) {
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

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
