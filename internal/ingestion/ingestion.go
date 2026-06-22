// Package ingestion handles ingestion of webhook events
package ingestion

import (
	"context"
	"log"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/queue"
)

// SubscriptionStore defines the data access interface needed by the ingestion service.
type SubscriptionStore interface {
	GetActiveSubscriptions(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error)
}

type IngestionService struct {
	store        SubscriptionStore
	queue        *queue.Broker
	idempotency  *IdempotencyStore
}

func NewIngestionService(store SubscriptionStore, queue *queue.Broker, idempotency *IdempotencyStore) *IngestionService {
	return &IngestionService{
		store:       store,
		queue:       queue,
		idempotency: idempotency,
	}
}

// IngestEvent looks up active subscriptions for the event and pushes a delivery job to the queue for each one.
func (s *IngestionService) IngestEvent(ctx context.Context, event domain.WebhookEvent) error {
	if s.idempotency.MarkSeen(event.ID) {
		log.Printf("duplicate event %s, skipping", event.ID)
		return nil
	}

	subscriptions, err := s.store.GetActiveSubscriptions(ctx, event.TenantID, event.EventType)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		job := domain.DeliveryJob{
			EventID:      event.ID,
			EventType:    event.EventType,
			TargetURL:    sub.TargetURL,
			SecretKey:    sub.SecretKey,
			Payload:      event.Payload,
			AttemptCount: 0,
		}

		if err := s.queue.Publish(ctx, job); err != nil {
			log.Printf("failed to publish delivery job for subscription %s: %v", sub.ID, err)
			continue
		}
	}

	return nil
}
