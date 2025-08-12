package subscription_service

import (
	"Weather-API-Application/internal/clients" // Припускаємо, що email-клієнт тут
	"Weather-API-Application/internal/config"  // Залишаємо для конфігурації email
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrSubscriptionExists = errors.New("subscription already exists")
	ErrNotFound           = errors.New("subscription not found")
	ErrAlreadyConfirmed   = errors.New("subscription already confirmed")
)

type SubscriptionService struct {
	repo        repository.SubscriptionRepository
	emailClient clients.EmailClient
	cfg         *config.Config
	mu          sync.Mutex
	routines    map[string]context.CancelFunc
}

func NewSubscriptionService(repo repository.SubscriptionRepository, emailClient clients.EmailClient, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		repo:        repo,
		emailClient: emailClient,
		cfg:         cfg,
		routines:    make(map[string]context.CancelFunc),
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, req *model.SubscriptionCreate) error {
	// 1. Отримуємо підписку з репозиторію
	sub, err := s.repo.GetByEmailAndCity(ctx, req.Email, req.City)

	// Ситуація: підписка вже існує
	if err == nil {
		if sub.Confirmed {
			// Знайдено і підтверджено -> повертаємо семантичну бізнес-помилку
			return ErrSubscriptionExists
		}

		// Знайдено, але не підтверджено -> оновлюємо токен і надсилаємо лист (успішний сценарій)
		newToken := createNewToken()
		if err := s.repo.UpdateToken(ctx, sub.ID, newToken); err != nil {
			// Це внутрішня помилка БД, "загортаємо" її для контексту
			return fmt.Errorf("failed to update subscription: %w", err)
		}

		// Надсилаємо лист
		if err := s.emailClient.SendEmail(ctx, sub.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}
		return nil // Успіх
	}

	// Ситуація: підписка не знайдена, створюємо нову
	if errors.Is(err, repository.ErrNotFound) { // Припускаємо, що репозиторій повертає свою помилку ErrNotFound
		newToken := createNewToken()
		newSub := &model.Subscription{
			Email:     req.Email,
			City:      req.City,
			Frequency: req.Frequency,
			Token:     newToken,
			Confirmed: false,
		}

		if err := s.repo.Create(ctx, newSub); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}

		if err := s.emailClient.SendEmail(ctx, newSub.Email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
			return fmt.Errorf("failed to send confirmation email: %w", err)
		}
		return nil // Успіх
	}

	// Будь-яка інша помилка з репозиторію є неочікуваною
	return fmt.Errorf("failed to check for subscription: %w", err)
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

func createNewToken() string {
	// ... ваша логіка створення токена
	return ""
}
