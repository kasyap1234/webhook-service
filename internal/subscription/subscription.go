package subscription

import (
	"context"
	"errors"

	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/security"
)

var ErrSubscriptionNotFound = errors.New("subscription not found or already inactive")

type SubscriptionService struct {
	repo *database.SubscriptionStore
}

func NewSubscriptionService(repo *database.SubscriptionStore) *SubscriptionService {
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
