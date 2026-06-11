package subscription

import (
	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/ingestion/subscription"
)

type SubscriptionService struct {
	repo database.SubscriptionRepo
}

func NewSubscriptionService(repo database.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) ActivateSubscription(subscription *domain.Subscription) error {
	subscription.IsActive = true
	
}