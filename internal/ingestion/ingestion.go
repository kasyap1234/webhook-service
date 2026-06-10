// Package ingestion handles ingestion of webhook events
package ingestion

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/zeromicro/go-zero/core/queue"
)

// / will receive requests from external serivce , receives raw event, looks up who is subscribed and packages it
//
// next will be the queue (holds the package safely)
//
// worker pulls from the queue and processes the event and fires it across the internet .
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

func (s *IngestionService) NewIngestionService(SubscriptionRepository SubscriptionStore, Queue *queue.Broker) *IngestionService {
	return &IngestionService{
		SubscriptionRepository: SubscriptionRepository,
		Queue:                  Queue,
	}
}

// push into the queue
func (s *IngestionService) ProcessEvent(event domain.WebhookEvent) error {
	err := s.Queue.Push(event)
	if err != nil {
		return err
	}

	return nil
}
