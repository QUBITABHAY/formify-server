package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitDB(databaseURL string) error {
	var err error
	DBPool, err = pgxpool.New(context.Background(), databaseURL)
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
