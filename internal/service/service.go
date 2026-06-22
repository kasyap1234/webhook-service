package service

import "github.com/kasyap1234/webhook-service/internal/subscription"

type Service struct {
	SubscriptionService *subscription.SubscriptionService
}

func NewService(subscriptionService *subscription.SubscriptionService) *Service {
	return &Service{
		SubscriptionService: subscriptionService,
	}
}


