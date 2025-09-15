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
	_ "Weather-API-Application/cmd/api/docs"
	"Weather-API-Application/internal/client"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/infrastructure/database"
	"Weather-API-Application/internal/infrastructure/repository"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/server"
	"Weather-API-Application/internal/server/handlers"
	"Weather-API-Application/internal/server/middleware"
	"Weather-API-Application/internal/services/scheduler_service"
	"Weather-API-Application/internal/services/subscription_service"
	"Weather-API-Application/internal/services/weather_service"
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

	// Initialize database
	db, err := database.NewPostgresDB(cfg.GetDSN())
	if err != nil {
		logger.Fatal(ctx, err)
	}

	// Initialize Email client
	emailClient := client.NewEmailClient(cfg)

	// Initialize repositories
	subscriptionRepository := repository.NewSubscriptionRepository(db)

	// Create services
	schedulerService := scheduler_service.NewSchedulerService(subscriptionRepository, *emailClient, cfg)
	subscriptionService := subscription_service.NewSubscriptionService(subscriptionRepository, *emailClient, cfg)
	weatherService := weather_service.NewService(cfg)

	// Create api
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

	// Run api
	srvr.Run(ctx)

}
