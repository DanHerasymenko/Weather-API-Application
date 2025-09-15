package scheduler_service

import (
	"Weather-API-Application/internal/client"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"Weather-API-Application/internal/services/email_service"
	"Weather-API-Application/internal/services/subscription_service"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Scheduler is responsible for spawning and managing weather update routines
// for all confirmed subscriptions.

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

// StartScheduler fetches all confirmed subscriptions from the database
// and starts a background goroutine for each one to send periodic weather updates.
func (s *SchedulerService) StartScheduler(ctx context.Context) error {
	subs, err := s.repo.ListConfirmed(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch confirmed subscriptions: %w", err)
	}

	for _, sub := range subs {

		// For each subscription, create a new own context with a cancel function
		subCtx, cancel := context.WithCancel(ctx)
		key := subscription_service.MakeKey(sub)

		s.mu.Lock()
		s.routines[key] = cancel
		s.mu.Unlock()

		// Start the routine for this particular subscription
		go s.StartRoutine(subCtx, sub)
	}

	logger.Info(ctx, fmt.Sprintf("Starting %d subscription routines", len(subs)))

	return nil
}

// startRoutine runs a background loop for a single subscription.
// It determines whether the updates should be sent hourly or daily,
// waits for the correct interval, then periodically calls sendUpdate.
// The routine stops when the provided context is cancelled.
func (s *SchedulerService) StartRoutine(ctx context.Context, sub *model.Subscription) {

	// Detect the interval based on the subscription frequency
	interval := time.Hour
	if strings.ToLower(sub.Frequency) == "daily" {
		interval = 24 * time.Hour

		// Calculate how long to wait until the next daily update
		now := time.Now()
		next := time.Date(
			now.Year(), now.Month(), now.Day(),
			s.cfg.DailyStartHour, 0, 0, 0,
			now.Location(),
		)

		// If current time is after the next scheduled time, add 24 hours
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Wait until the next scheduled time
		select {
		case <-time.After(time.Until(next)):
			// continue to the loop with proper start time
		case <-ctx.Done():
			logger.Info(ctx, fmt.Sprintf("Routine cancelled before first run: %s - %s", sub.Email, sub.City))
			return
		}
	} else {
		interval = time.Hour
	}

	// Create a ticker to send updates with the specified interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, fmt.Sprintf("Stopping routine for %s - %s", sub.Email, sub.City))
			return

		case <-ticker.C:
			logger.Info(ctx, fmt.Sprintf("Attempting to send update to %s for %s", sub.Email, sub.City))
			if err := email_service.SendUpdate(ctx, s.cfg.WeatherApiKey, sub, s.emailClient); err != nil {
				logger.Error(ctx, err)
			}
			logger.Info(ctx, fmt.Sprintf("Weather update sent to %s for city %s", sub.Email, sub.City))
		}
	}
}
