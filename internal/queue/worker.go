package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/kasyap1234/webhook-service/internal/domain"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

// DeliveryHandler processes a delivered webhook job and returns an error if it should be retried.
type DeliveryHandler func(ctx context.Context, job domain.DeliveryJob) error

type Worker struct {
	consumer *rabbitmq.Consumer
	handler  DeliveryHandler
}

func NewWorker(conn *rabbitmq.Conn, handler DeliveryHandler) (*Worker, error) {
	if handler == nil {
		return nil, errors.New("queue: delivery handler is required")
	}

	consumer, err := rabbitmq.NewConsumer(conn, "webhooks",
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName("webhooks"),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsRoutingKey("#"),
		rabbitmq.WithConsumerOptionsLogging)
	if err != nil {
		return nil, err
	}

	return &Worker{
		consumer: consumer,
		handler:  handler,
	}, nil
}

// deliver is the internal rabbitmq handler that unmarshals delivery jobs and delegates.
func (w *Worker) deliver(d rabbitmq.Delivery) rabbitmq.Action {
	var job domain.DeliveryJob
	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("failed to unmarshal delivery job: %v", err)
		return rabbitmq.NackDiscard
	}

	if err := w.handler(context.Background(), job); err != nil {
		log.Printf("delivery handler failed for event %s: %v (will requeue)", job.EventID, err)
		return rabbitmq.NackRequeue
	}

	return rabbitmq.Ack
}

// Start begins consuming messages from the queue. This is a blocking call — run it in a goroutine.
func (w *Worker) Start() error {
	return w.consumer.Run(w.deliver)
}

// Close gracefully shuts down the consumer.
func (w *Worker) Close() {
	w.consumer.Close()
}
