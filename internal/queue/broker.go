package queue

import (
	"context"
	"encoding/json"

	"github.com/kasyap1234/webhook-service/internal/domain"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type Broker struct {
	publisher *rabbitmq.Publisher
}

func NewBroker(conn *rabbitmq.Conn) (*Broker, error) {
	publisher, err := rabbitmq.NewPublisher(conn, rabbitmq.WithPublisherOptionsExchangeName("webhooks"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsLogging)
	if err != nil {
		return nil, err
	}
	return &Broker{publisher: publisher}, nil
}

func (b *Broker) Publish(ctx context.Context, job domain.DeliveryJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return b.publisher.PublishWithContext(ctx, body, []string{job.EventType}, rabbitmq.WithPublishOptionsContentType("application/json"), rabbitmq.WithPublishOptionsExchange("webhooks"), rabbitmq.WithPublishOptionsPersistentDelivery)
}

func (b *Broker) Close() error {
	b.publisher.Close()
	return nil
}
