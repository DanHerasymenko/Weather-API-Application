package database

import (
	"Weather-API-Application/internal/logger"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Postgres *pgxpool.Pool
}

// NewPostgresClient creates a new Postgres client
func NewPostgresDB(ctx context.Context, dsn string) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// check connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	logger.Info(ctx, "Postgres ping successful")

	return &PostgresDB{Postgres: pool}, nil
}
