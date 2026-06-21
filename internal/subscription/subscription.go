package subscription

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/security"
)

type SubscriptionService struct {
	repo *database.SubscriptionStore
}

func NewSubscriptionService(repo *database.SubscriptionStore) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) ActivateSubscription(ctx context.Context, payload string, tenantID, eventType, targetURL string) (string, error) {
	// check if subscription already exists and is active
	subscription, err := s.repo.GetSubscription(ctx, tenantID, eventType, targetURL)
	if err != nil {
		return "", err
	}

	if subscription != nil && subscription.IsActive {
		return subscription.SecretKey, nil
	}

	// generate a secure key for the subscription
	secretKey, err := security.GenerateSecureKey()
	if err != nil {
		return "", err
	}

	// create the subscription with the generated secret key
	if err := s.repo.Subscribe(ctx, tenantID, eventType, targetURL, secretKey); err != nil {
		return "", err
	}

	return secretKey, nil
}

func (s *SubscriptionService) DeactivateSubscription(ctx context.Context, tenantID, subscriptionID string) error {
	return s.repo.Unsubscribe(ctx, tenantID, subscriptionID)
}
