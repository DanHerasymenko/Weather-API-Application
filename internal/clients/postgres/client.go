package postgres

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "github.com/pressly/goose/v3"
)

type Client struct {
	Postgres *pgxpool.Pool
}

// NewPostgresClient creates a new Postgres client
func NewPostgresClient(ctx context.Context, cfg *config.Config) (*Client, error) {

	// create a new Postgres connection pool
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresContainerHost,
		cfg.PostgresContainerPort,
		cfg.PostgresDB,
	)

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

	// run migrations if the RUN_MIGRATIONS env variable is set to true
	if cfg.RunMigrations {
		if err := runMigrations(ctx, dsn); err != nil {
			return nil, err
		}
		logger.Info(ctx, "Postgres migrations successful")
	} else {
		logger.Info(ctx, "Postgres migrations skipped")
	}

	return &Client{Postgres: pool}, nil
}

// runMigrations runs the database migrations using goose
func runMigrations(ctx context.Context, dsn string) error {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer func() {
		if cl := db.Close(); cl != nil {
			logger.Info(ctx, fmt.Sprintf("failed to close migration db connection: %v", cl))
		}
	}()

	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return nil
}
