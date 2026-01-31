package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitDB() error {
	var err error
	DBPool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	return DBPool.Ping(context.Background())
}

func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
	}
}
