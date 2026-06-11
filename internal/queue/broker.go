package queue

import (
	"context"

	"github.com/kasyap1234/webhook-service/internal/domain"
)

type Broker struct{}

func NewBroker() *Broker {
	return &Broker{}
}

type JobPublisher interface {
	Publish(ctx context.Context, job domain.DeliveryJob) error
}

func (b *Broker) Publish(event domain.WebhookEvent) error {
	return nil
}
