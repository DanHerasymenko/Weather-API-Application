// @title           Weather Forecast API
// @version         1.0
// @description     Weather API application that allows users to subscribe to weather updates for their city.

package main

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/server"
	"Weather-API-Application/internal/services"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to load config: %w", err))
	}
	logger.Info(ctx, "Config loaded")

	// Create clients (Postgres client)
	clnts, err := clients.NewClients(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to create clients: %w", err))
	}

	// Create services
	srvc := services.NewServices(cfg, clnts)

	// Create server
	srvr := server.NewServer(cfg)

	// Register middlewares
	mdlwrs := middleware.NewMiddlewares(cfg, clnts)
}
