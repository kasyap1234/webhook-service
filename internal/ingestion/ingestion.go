// Package ingestion handles ingestion of webhook events
package ingestion

import (
	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/zeromicro/go-zero/core/queue"
)

// / will receive requests from external serivce , receives raw event, looks up who is subscribed and packages it
//
// next will be the queue (holds the package safely)
//
// worker pulls from the queue and processes the event and fires it across the internet .
type IngestionHandler struct {
	SubscriptionRepository domain.SubscriptionRepository
	Queue                  *queue.Queue
}
