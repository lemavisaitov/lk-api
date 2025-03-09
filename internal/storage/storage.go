package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetConnect(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
