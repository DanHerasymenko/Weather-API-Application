package scheduler_service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"Weather-API-Application/internal/client"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"Weather-API-Application/internal/services/subscription_service"
)

// SchedulerService manages background weather update routines for confirmed subscriptions.
type SchedulerService struct {
	repo        repository.SubscriptionRepository
	emailClient client.EmailClient
	cfg         *config.Config
	mu          sync.Mutex
	routines    map[string]context.CancelFunc
}

func NewSchedulerService(repo repository.SubscriptionRepository, emailClient client.EmailClient, cfg *config.Config) *SchedulerService {
	return &SchedulerService{
		repo:        repo,
		emailClient: emailClient,
		cfg:         cfg,
		routines:    make(map[string]context.CancelFunc),
	}
}

// StartScheduler starts routines for all confirmed subscriptions.
func (s *SchedulerService) StartScheduler(ctx context.Context) error {
	subs, err := s.repo.ListConfirmed(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch confirmed subscriptions: %w", err)
	}

	for _, sub := range subs {
		// Create a context for this subscription
		subCtx, cancel := context.WithCancel(ctx)
		key := subscription_service.MakeKey(sub)

		s.mu.Lock()
		s.routines[key] = cancel
		s.mu.Unlock()

		go s.StartRoutine(subCtx, sub)
	}

	logger.Info(ctx, "Starting subscription routines",
		slog.Int("count", len(subs)))
	return nil
}

// StartRoutine runs periodic updates for a single subscription until the context is cancelled.
func (s *SchedulerService) StartRoutine(ctx context.Context, sub *model.Subscription) {
	// Determine interval
	interval := time.Hour
	if strings.ToLower(sub.Frequency) == "daily" {
		interval = 24 * time.Hour

		// Wait until next scheduled daily start
		now := time.Now()
		next := time.Date(
			now.Year(), now.Month(), now.Day(),
			s.cfg.DailyStartHour, 0, 0, 0,
			now.Location(),
		)
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			logger.Info(ctx, "Routine cancelled before first run",
				slog.String("email", sub.Email),
				slog.String("city", sub.City))
			return
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "Stopping routine",
				slog.String("email", sub.Email),
				slog.String("city", sub.City))
			return
		case <-ticker.C:
			logger.Info(ctx, "Attempting to send update",
				slog.String("email", sub.Email),
				slog.String("city", sub.City))
			if err := client.SendUpdate(ctx, s.cfg.WeatherApiKey, sub, s.emailClient); err != nil {
				logger.Error(ctx, err,
					slog.String("email", sub.Email),
					slog.String("city", sub.City))
			} else {
				logger.Info(ctx, "Weather update sent",
					slog.String("email", sub.Email),
					slog.String("city", sub.City))
			}
		}
	}
}
