package queue

import "github.com/kasyap1234/webhook-service/internal/domain"

type Broker struct{}

func NewBroker() *Broker {
	return &Broker{}
}

func (b *Broker) Push(event domain.WebhookEvent) error {
	return nil
}
