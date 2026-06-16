package subscription

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/database"
)

type SubscriptionService struct {
	repo *database.SubscriptionRepo
}

func NewSubscriptionService(repo *database.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) ActivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) error {
	subscription, err := s.repo.GetSubscription(ctx, tenantID, eventType, targetURL)
	if err != nil {
		return err
	}

	if subscription != nil && subscription.IsActive {
		return nil
	}

	return s.repo.Subscribe(ctx, tenantID, eventType, targetURL)
}

func (s *SubscriptionService) DeactivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) error {
	return s.repo.Unsubscribe(ctx, tenantID, eventType, targetURL)
}
