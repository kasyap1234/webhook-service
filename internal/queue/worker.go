package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/kasyap1234/webhook-service/internal/domain"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

// DeliveryHandler processes a delivered webhook job and returns the result or an error.
type DeliveryHandler func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error)

// DeliveryLogger persists delivery attempt results.
type DeliveryLogger interface {
	LogDelivery(ctx context.Context, log *domain.DeliveryLog) error
}

type Worker struct {
	consumer       *rabbitmq.Consumer
	handler        DeliveryHandler
	deliveryLogger DeliveryLogger
}

func NewWorker(conn *rabbitmq.Conn, handler DeliveryHandler, deliveryLogger DeliveryLogger) (*Worker, error) {
	if handler == nil {
		return nil, errors.New("queue: delivery handler is required")
	}

	consumer, err := rabbitmq.NewConsumer(conn, "webhooks",
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName("webhooks"),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsRoutingKey("#"),
		rabbitmq.WithConsumerOptionsLogging,
		rabbitmq.WithConsumerOptionsConcurrency(10),
		rabbitmq.WithConsumerOptionsQOSPrefetch(10))
	if err != nil {
		return nil, err
	}

	return &Worker{
		consumer:       consumer,
		handler:        handler,
		deliveryLogger: deliveryLogger,
	}, nil
}

// deliver is the internal rabbitmq handler that unmarshals delivery jobs and delegates.
func (w *Worker) deliver(d rabbitmq.Delivery) rabbitmq.Action {
	ctx := context.Background()

	var job domain.DeliveryJob
	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("failed to unmarshal delivery job: %v", err)
		return rabbitmq.NackDiscard
	}

	start := time.Now()
	result, err := w.handler(ctx, job)
	durationMs := int(time.Since(start).Milliseconds())

	entry := &domain.DeliveryLog{
		EventID:        job.EventID,
		SubscriptionID: job.SubscriptionID,
		TenantID:       job.TenantID,
		EventType:      job.EventType,
		TargetURL:      job.TargetURL,
		AttemptNumber:  job.AttemptCount + 1,
		DurationMs:     durationMs,
	}

	if err != nil {
		entry.ErrorMessage = err.Error()
		job.AttemptCount++

		if job.AttemptCount >= domain.MaxDeliveryAttempts {
			entry.StatusCode = result.StatusCode
			entry.ResponseBody = result.ResponseBody
			entry.Status = "failed"
			w.saveLog(ctx, entry)
			log.Printf("delivery failed for event %s after %d attempts, discarding: %v", job.EventID, job.AttemptCount, err)
			return rabbitmq.NackDiscard
		}

		entry.StatusCode = result.StatusCode
		entry.ResponseBody = result.ResponseBody
		entry.Status = "attempt_failed"
		w.saveLog(ctx, entry)
		log.Printf("delivery handler failed for event %s (attempt %d/%d): %v", job.EventID, job.AttemptCount, domain.MaxDeliveryAttempts, err)
		return rabbitmq.NackRequeue
	}

	entry.StatusCode = result.StatusCode
	entry.ResponseBody = result.ResponseBody
	entry.Status = "success"
	w.saveLog(ctx, entry)

	return rabbitmq.Ack
}

func (w *Worker) saveLog(ctx context.Context, entry *domain.DeliveryLog) {
	if w.deliveryLogger == nil {
		return
	}
	if err := w.deliveryLogger.LogDelivery(ctx, entry); err != nil {
		log.Printf("failed to save delivery log for event %s: %v", entry.EventID, err)
	}
}

// Start begins consuming messages from the queue. This is a blocking call — run it in a goroutine.
func (w *Worker) Start() error {
	return w.consumer.Run(w.deliver)
}

// Close gracefully shuts down the consumer.
func (w *Worker) Close() {
	w.consumer.Close()
}
