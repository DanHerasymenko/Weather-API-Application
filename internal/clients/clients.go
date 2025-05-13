package clients

import (
	"Weather-API-Application/internal/clients/postgres"
	"Weather-API-Application/internal/config"
	"context"
)

type Clients struct {
	PostgresClnt *postgres.Client
}

func NewClients(ctx context.Context, cfg *config.Config) (*Clients, error) {

	postgresClient, err := postgres.NewPostgresClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	clients := &Clients{
		PostgresClnt: postgresClient,
	}

	return clients, nil

}
