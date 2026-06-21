package handlers

import "github.com/kasyap1234/webhook-service/internal/subscription"

type Handler struct {
	SubscriptionHandler *SubscriptionHandler
}

func NewHandler(subscriptionService *subscription.SubscriptionService) *Handler {
	return &Handler{
		SubscriptionHandler: NewSubscriptionHandler(subscriptionService),
	}
}
