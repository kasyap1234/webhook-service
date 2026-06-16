package subscription

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/security"
)

type SubscriptionService struct {
	repo *database.SubscriptionRepo
}

func NewSubscriptionService(repo *database.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) ActivateSubscription(ctx context.Context,payload string, tenantID, eventType, targetURL string) error {
	// check if subscription already exists and is active
	subscription, err := s.repo.GetSubscription(ctx, tenantID, eventType, targetURL)
	secretKey,err  :=security.GenerateSecureKey()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if subscription != nil && subscription.IsActive {
		return nil
	}
	// if not active , activate it but first secure it
	// first generate a secure key for the subscription
	secureKey,err :=security.GenerateSignature([]byte(payload),secretKey)
	if err !=nil{
		return err 
	}
	// secure the subscription with the generated key
	return s.repo.Subscribe(ctx, tenantID, eventType, targetURL, secureKey)
}

func (s *SubscriptionService) DeactivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) error {
	return s.repo.Unsubscribe(ctx, tenantID, eventType, targetURL)
}
