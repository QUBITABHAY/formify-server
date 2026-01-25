package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitDB() error {
	var err error
	DBPool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	err = DBPool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
		log.Println("Database connection closed")
	}
}
