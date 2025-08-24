package subscription_service

import (
	"Weather-API-Application/internal/client"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"context"
	"errors"
	"fmt"
	"sync"
)

type SubscriptionService struct {
	repo        repository.SubscriptionRepository
	emailClient client.EmailClient
	cfg         *config.Config
	mu          sync.Mutex
	routines    map[string]context.CancelFunc
}

func NewSubscriptionService(repo repository.SubscriptionRepository, emailClient client.EmailClient, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		repo:        repo,
		emailClient: emailClient,
		cfg:         cfg,
		routines:    make(map[string]context.CancelFunc),
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, req *model.Subscription) error {

	rowExists, confirmed, err := s.repo.CheckConfirmation(ctx, req)
	if err != nil {
		return fmt.Errorf("check confirmation: %w", err)
	}

	// 1) Нема підписки -> створюємо і відправляємо лист
	if !rowExists {

		token := CreateNewToken()
		sub := &model.Subscription{
			Email:     req.Email,
			City:      req.City,
			Frequency: req.Frequency,
			Token:     token,
			Confirmed: false,
		}

		err := s.repo.Create(ctx, sub)
		if err != nil {
			return ErrFailedToCreateSubscription
		}

		err = s.emailClient.SendEmail(ctx, sub.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, token))
		if err != nil {
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}

		return nil
	}

	// 2) Є, але не підтверджена -> оновлюємо токен і шлемо лист
	if !confirmed {
		token := CreateNewToken()

		// Якщо в тебе немає ID, зроби метод, який оновлює токен по (email, city)
		err := s.repo.UpdateTokenByEmailCity(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update subscription token: %w", err)
		}

		err = s.emailClient.SendEmail(ctx, req.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, token))
		if err != nil {
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}
		return nil
	}

	// 3) Є і підтверджена -> бізнес-правило: вважаємо помилкою
	return ErrSubscriptionExists
}

func (s *SubscriptionService) ConfirmSubscription(ctx context.Context, token string) error {
	sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Транслюємо помилку репозиторію в нашу бізнес-помилку
			return ErrNotFound
		}
		return fmt.Errorf("failed to scan subscription: %w", err) // Інша помилка БД
	}

	if sub.Confirmed {
		return ErrAlreadyConfirmed
	}

	// Позначаємо підписку як підтверджену через репозиторій
	if err := s.repo.SetConfirmed(ctx, sub.ID); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Запускаємо фонову задачу
	go s.startRoutine(ctx, sub)

	return nil // Успіх
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	// 1. Отримуємо підписку, щоб знати, яку рутину зупинити
	sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// 2. Видаляємо підписку через репозиторій
	if err := s.repo.DeleteByToken(ctx, token); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	// 3. Зупиняємо рутину
	key := s.MakeKey(sub)
	s.mu.Lock()
	if cancel, ok := s.routines[key]; ok {
		cancel()
		delete(s.routines, key)
	}
	s.mu.Unlock()

	return nil // Успіх
}

func (s *SubscriptionService) fetchConfirmedSubscriptions(ctx context.Context) ([]*model.Subscription, error) {
	// Уся логіка запитів до БД знаходиться в репозиторії
	return s.repo.ListConfirmed(ctx)
}

// Цей метод більше не потрібен у сервісі, оскільки його логіка
// тепер повністю інкапсульована в s.repo.Create()
/*
func (s *SubscriptionService) createNewSubscription(...) error { ... }
*/

// Приватні методи, які є частиною бізнес-логіки (управління рутинами), залишаються тут
func (s *SubscriptionService) startRoutine(ctx context.Context, sub *model.Subscription) {
	// ... ваша логіка запуску
}

func (s *SubscriptionService) MakeKey(sub *model.Subscription) string {
	// ... ваша логіка створення ключа
	return ""
}
