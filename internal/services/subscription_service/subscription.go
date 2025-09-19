package subscription_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"

	"Weather-API-Application/internal/client"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
)

type Scheduler interface {
	StartFor(ctx context.Context, sub *model.Subscription)
	StopFor(sub *model.Subscription)
}

type SubscriptionService struct {
	repo        repository.SubscriptionRepository
	emailClient client.EmailClient
	cfg         *config.Config
	scheduler   Scheduler
	mu          sync.Mutex
}

func NewSubscriptionService(repo repository.SubscriptionRepository, emailClient client.EmailClient, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		repo:        repo,
		emailClient: emailClient,
		cfg:         cfg,
	}
}

func (s *SubscriptionService) WithScheduler(scheduler Scheduler) *SubscriptionService {
	s.scheduler = scheduler
	return s
}

// Subscribe creates a new subscription or updates a pending one and sends a confirmation email.
func (s *SubscriptionService) Subscribe(ctx context.Context, req *model.Subscription) error {
	rowExists, confirmed, err := s.repo.CheckConfirmation(ctx, req)
	if err != nil {
		return fmt.Errorf("check confirmation: %w", err)
	}

	// 1) No subscription -> create and send confirmation email
	if !rowExists {
		token := createNewToken()
		sub := &model.Subscription{
			Email:     req.Email,
			City:      req.City,
			Frequency: req.Frequency,
			Token:     token,
			Confirmed: false,
		}

		if err := s.repo.Create(ctx, sub); err != nil {
			return ErrFailedToCreateSubscription
		}

		if err := s.emailClient.SendEmail(ctx, sub.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, token)); err != nil {
			logger.Error(ctx, err,
				slog.String("email", sub.Email),
				slog.String("city", sub.City))
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}
		logger.Info(ctx, "Confirmation email sent",
			slog.String("email", sub.Email),
			slog.String("city", sub.City))
		return nil
	}

	// 2) Exists but not confirmed -> update token and resend confirmation
	if !confirmed {
		token := createNewToken()
		req.Token = token
		if err := s.repo.UpdateTokenByEmailCity(ctx, req); err != nil {
			return fmt.Errorf("failed to update subscription token: %w", err)
		}

		if err := s.emailClient.SendEmail(ctx, req.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, token)); err != nil {
			logger.Error(ctx, err,
				slog.String("email", req.Email),
				slog.String("city", req.City))
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}
		logger.Info(ctx, "Confirmation email resent",
			slog.String("email", req.Email),
			slog.String("city", req.City))
		return nil
	}

	// 3) Exists and confirmed -> business rule: treat as error
	return ErrSubscriptionExists
}

// ConfirmSubscription confirms subscription by token.
func (s *SubscriptionService) ConfirmSubscription(ctx context.Context, token string) (*model.Subscription, error) {
	subId, sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan subscription: %w", err)
	}

	if sub.Confirmed {
		return nil, ErrAlreadyConfirmed
	}

	if err := s.repo.SetConfirmed(ctx, subId); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info(ctx, "Subscription confirmed",
		slog.String("email", sub.Email),
		slog.String("city", sub.City))

	if s.scheduler != nil {
		s.scheduler.StartFor(ctx, sub)
	}
	return sub, nil
}

// Unsubscribe removes subscription by token and stops its routine if running.
func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	_, sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to scan subscription: %w", err)
	}

	if err := s.repo.DeleteByToken(ctx, token); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if s.scheduler != nil {
		s.scheduler.StopFor(sub)
	}

	logger.Info(ctx, "Subscription unsubscribed",
		slog.String("email", sub.Email),
		slog.String("city", sub.City))
	return nil
}

func (s *SubscriptionService) fetchConfirmedSubscriptions(ctx context.Context) ([]*model.Subscription, error) {
	return s.repo.ListConfirmed(ctx)
}

func MakeKey(sub *model.Subscription) string {
	return fmt.Sprintf("%s|%s", sub.Email, strings.ToLower(sub.City))
}

func createNewToken() string {
	return uuid.New().String()
}
