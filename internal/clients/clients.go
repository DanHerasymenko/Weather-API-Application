package clients

import (
	"Weather-API-Application/internal/clients/email"
	"Weather-API-Application/internal/clients/postgres"
	"Weather-API-Application/internal/config"
	"context"
)

type Clients struct {
	PostgresClnt *postgres.Client
	EmailClnt    email.Client
}

func NewClients(ctx context.Context, cfg *config.Config) (*Clients, error) {

	postgresClient, err := postgres.NewPostgresClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	emailClient := email.NewSMTPClient(
		cfg.From,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	clients := &Clients{
		PostgresClnt: postgresClient,
		EmailClnt:    emailClient,
	}

	return clients, nil

}
