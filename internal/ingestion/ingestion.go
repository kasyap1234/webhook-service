// Package ingestion handles ingestion of webhook events
package ingestion

import (
	"context"
	"log"

	"github.com/kasyap1234/webhook-service/internal/domain"
)

// SubscriptionFinder retrieves active subscriptions for a given tenant and event type.
type SubscriptionFinder interface {
	GetActiveSubscriptions(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error)
}

// EventPersister persists incoming webhook events.
type EventPersister interface {
	CreateEvent(ctx context.Context, event *domain.WebhookEvent) error
}

// Publisher publishes delivery jobs to a message queue.
type Publisher interface {
	Publish(ctx context.Context, job domain.DeliveryJob) error
}

type IngestionService struct {
	subscriptions SubscriptionFinder
	events        EventPersister
	queue         Publisher
	idempotency   *IdempotencyStore
}

func NewIngestionService(subscriptions SubscriptionFinder, events EventPersister, queue Publisher, idempotency *IdempotencyStore) *IngestionService {
	return &IngestionService{
		subscriptions: subscriptions,
		events:        events,
		queue:         queue,
		idempotency:   idempotency,
	}
}

// IngestEvent looks up active subscriptions for the event and pushes a delivery job to the queue for each one.
func (s *IngestionService) IngestEvent(ctx context.Context, event domain.WebhookEvent) error {
	if s.idempotency.MarkSeen(event.ID) {
		log.Printf("duplicate event %s, skipping", event.ID)
		return nil
	}

	if err := s.events.CreateEvent(ctx, &event); err != nil {
		return err
	}

	subscriptions, err := s.subscriptions.GetActiveSubscriptions(ctx, event.TenantID, event.EventType)
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
