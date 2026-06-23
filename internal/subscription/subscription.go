package subscription

import (
	"context"
	"errors"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/security"
)

var ErrSubscriptionNotFound = errors.New("subscription not found or already inactive")

type SubscriptionRepository interface {
	Subscribe(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error
	GetSubscription(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error)
	Unsubscribe(ctx context.Context, tenantID, subscriptionID string) error
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) ActivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) (string, error) {
	secretKey, err := security.GenerateSecureKey()
	if err != nil {
		return "", err
	}

	if err := s.repo.Subscribe(ctx, tenantID, eventType, targetURL, secretKey); err != nil {
		return "", err
	}

	return secretKey, nil
}

func (s *SubscriptionService) DeactivateSubscription(ctx context.Context, tenantID, subscriptionID string) error {
	subscription, err := s.repo.GetSubscription(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	if subscription == nil || !subscription.IsActive {
		return ErrSubscriptionNotFound
	}

	return s.repo.Unsubscribe(ctx, tenantID, subscriptionID)
}
