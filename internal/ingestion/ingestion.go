// Package ingestion handles ingestion of webhook events
package ingestion

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/queue"
)

type IngestionService struct {
	SubscriptionStore SubscriptionStore
	Queue             *queue.Broker
}

type SubscriptionStore interface {
	GetActiveSubscriptions(tenantID string, eventType string) ([]domain.Subscription, error)
}

func NewIngestionService(SubscriptionStore SubscriptionStore, Queue *queue.Broker) *IngestionService {
	return &IngestionService{
		SubscriptionStore: SubscriptionStore,
		Queue:             Queue,
	}
}

func (s *IngestionService) PushDeliveryJob(ctx context.Context, job domain.DeliveryJob) error {
	return s.Queue.Publish(ctx, job)
}
