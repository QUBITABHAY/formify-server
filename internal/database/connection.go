// Package database provides database connection lifecycle helpers.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//nolint:gochecknoglobals // Application-wide shared DB pool.
var DBPool *pgxpool.Pool

func InitDB(databaseURL string) error {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}

	// Maintain warm connections to eliminate cold start TLS handshakes with Neon
	cfg.MinConns = 3
	cfg.MaxConns = 20
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	var poolErr error
	DBPool, poolErr = pgxpool.NewWithConfig(context.Background(), cfg)
	if poolErr != nil {
		return fmt.Errorf("init db pool: %w", poolErr)
	}
	return DBPool.Ping(context.Background())
}

func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
	}
}
