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

type JobPublisher interface {
	Publish(ctx context.Context, job domain.DeliveryJob) error
}

type SubscriptionStore interface {
	GetActiveSubscriptions(tenantID string, eventType string) ([]domain.Subscription, error)
}

func (s *IngestionService) NewIngestionService(SubscriptionStore SubscriptionStore, Queue *queue.Broker) *IngestionService {
	return &IngestionService{
		SubscriptionStore: SubscriptionStore,
		Queue:             Queue,
	}
}

func (s *IngestionService) ProcessEvent(event domain.WebhookEvent) error {
	err := s.Queue.Push(event)
	if err != nil {
		return err
	}

	return nil
}
