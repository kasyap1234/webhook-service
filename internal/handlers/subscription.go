package handlers

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/subscription"
)

type SubscriptionHandler struct {
	Service *subscription.SubscriptionService
}

func NewSubscriptionHandler(service *subscription.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		Service: service,
	}
}

func (h *SubscriptionHandler) ActivateSubscription(ctx context.Context, payload string, tenantID, eventType, targetURL string) error {
	return h.Service.ActivateSubscription(ctx, payload, tenantID, eventType, targetURL)
}
