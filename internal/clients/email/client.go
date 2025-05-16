package email

import (
	"Weather-API-Application/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Postgres *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, cfg *config.Config) (*Client, error) {
