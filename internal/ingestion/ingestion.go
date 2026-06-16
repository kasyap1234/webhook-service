// Package ingestion handles ingestion of webhook events
package ingestion

import (
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

func (s *IngestionService) NewIngestionService(SubscriptionStore SubscriptionStore, Queue *queue.Broker) *IngestionService {
	return &IngestionService{
		SubscriptionStore: SubscriptionStore,
		Queue:             Queue,
	}
}

func (s*IngestionService)PushDeliveryJob(job domain.DeliveryJob,)