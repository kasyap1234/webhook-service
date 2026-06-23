//go:build integration

package queue_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/queue"
	"github.com/kasyap1234/webhook-service/internal/testutil"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestBroker_Publish_Success(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	broker, err := queue.NewBroker(deps.RabbitConn)
	if err != nil {
		t.Fatalf("NewBroker failed: %v", err)
	}
	defer broker.Close()

	ch, err := deps.AmqpConn.Channel()
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("webhooks", "topic", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare exchange: %v", err)
	}

	queueName, err := ch.QueueDeclare("test-consumer", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare queue: %v", err)
	}

	err = ch.QueueBind(queueName.Name, "payment.completed", "webhooks", false, nil)
	if err != nil {
		t.Fatalf("failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(queueName.Name, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to consume: %v", err)
	}

	job := domain.DeliveryJob{
		EventID:   "evt-broker-1",
		EventType: "payment.completed",
		TargetURL: "https://example.com",
		SecretKey: "secret",
		Payload:   map[string]string{"amount": "100"},
	}

	err = broker.Publish(ctx, job)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg, ok := <-msgs:
		if !ok {
			t.Fatal("channel closed before receiving message")
		}

		var received domain.DeliveryJob
		if err := json.Unmarshal(msg.Body, &received); err != nil {
			t.Fatalf("failed to unmarshal received message: %v", err)
		}
		if received.EventID != "evt-broker-1" {
			t.Errorf("expected EventID=evt-broker-1, got %s", received.EventID)
		}
		if received.EventType != "payment.completed" {
			t.Errorf("expected EventType=payment.completed, got %s", received.EventType)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestBroker_Publish_RoutingKey(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	broker, err := queue.NewBroker(deps.RabbitConn)
	if err != nil {
		t.Fatalf("NewBroker failed: %v", err)
	}
	defer broker.Close()

	ch, err := deps.AmqpConn.Channel()
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("webhooks", "topic", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare exchange: %v", err)
	}

	queueName, err := ch.QueueDeclare("test-routing", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare queue: %v", err)
	}

	err = ch.QueueBind(queueName.Name, "order.created", "webhooks", false, nil)
	if err != nil {
		t.Fatalf("failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(queueName.Name, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("failed to consume: %v", err)
	}

	job := domain.DeliveryJob{
		EventID:   "evt-routing-1",
		EventType: "order.created",
		TargetURL: "https://example.com",
		SecretKey: "secret",
		Payload:   "data",
	}

	err = broker.Publish(ctx, job)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg, ok := <-msgs:
		if !ok {
			t.Fatal("channel closed")
		}
		if msg.RoutingKey != "order.created" {
			t.Errorf("expected routing key order.created, got %s", msg.RoutingKey)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestWorker_HandleDelivery_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	var handledJobs []domain.DeliveryJob
	handler := func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
		handledJobs = append(handledJobs, job)
		return domain.DeliveryJobResult{StatusCode: 200, ResponseBody: "ok"}, nil
	}

	worker, err := queue.NewWorker(deps.RabbitConn, handler, nil)
	if err != nil {
		t.Fatalf("NewWorker failed: %v", err)
	}
	defer worker.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start()
	}()

	time.Sleep(2 * time.Second)

	broker, err := queue.NewBroker(deps.RabbitConn)
	if err != nil {
		t.Fatalf("NewBroker failed: %v", err)
	}
	defer broker.Close()

	job := domain.DeliveryJob{
		EventID:   "evt-worker-1",
		EventType: "#",
		TargetURL: "https://example.com",
		SecretKey: "secret",
		Payload:   map[string]string{"test": "value"},
	}

	err = broker.Publish(ctx, job)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	time.Sleep(3 * time.Second)

	if len(handledJobs) < 1 {
		t.Errorf("expected at least 1 handled job, got %d", len(handledJobs))
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("worker returned error: %v", err)
		}
	default:
	}
}

func TestWorker_HandleDelivery_Retry(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	attemptCount := 0
	handler := func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
		attemptCount++
		if attemptCount < 3 {
			return domain.DeliveryJobResult{StatusCode: 500, ResponseBody: "error"}, &amqp.Error{
				Code:   500,
				Reason: "test error",
			}
		}
		return domain.DeliveryJobResult{StatusCode: 200, ResponseBody: "ok"}, nil
	}

	worker, err := queue.NewWorker(deps.RabbitConn, handler, nil)
	if err != nil {
		t.Fatalf("NewWorker failed: %v", err)
	}
	defer worker.Close()

	go worker.Start()
	time.Sleep(2 * time.Second)

	broker, err := queue.NewBroker(deps.RabbitConn)
	if err != nil {
		t.Fatalf("NewBroker failed: %v", err)
	}
	defer broker.Close()

	job := domain.DeliveryJob{
		EventID:      "evt-retry-1",
		EventType:    "#",
		TargetURL:    "https://example.com",
		SecretKey:    "secret",
		Payload:      "data",
		AttemptCount: 0,
	}

	err = broker.Publish(ctx, job)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	time.Sleep(5 * time.Second)

	if attemptCount < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attemptCount)
	}
}
