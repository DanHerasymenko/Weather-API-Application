// @title Weather Forecast API
// @version 1.0.0
// @description Weather API application that allows users to subscribe to weather updates for their city.
// @BasePath /api
// @schemes http https

// @tag.name weather
// @tag.description Weather forecast operations

// @tag.name subscription
// @tag.description Subscription management operations
package main

import (
	_ "Weather-API-Application/cmd/server/docs"
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/server"
	"Weather-API-Application/internal/server/handlers"
	"Weather-API-Application/internal/server/middleware"
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

	// Create clients
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

	// Create handlers
	hdlrs := handlers.NewHandlers(cfg, srvc, mdlwrs)
	hdlrs.RegisterRoutes(srvr.Router)

	// Start weather update scheduler (send emails by goroutines)
	if err := srvc.Subscription.StartScheduler(ctx); err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to start subscription scheduler: %w", err))
	}

	// Run server
	srvr.Run(ctx)

}
