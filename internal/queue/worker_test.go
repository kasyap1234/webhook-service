package queue

import (
	"context"
	"errors"
	"testing"

	"github.com/kasyap1234/webhook-service/internal/domain"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type mockDeliveryHandler struct {
	fn    func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error)
	calls int
}

func (m *mockDeliveryHandler) handle(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
	m.calls++
	if m.fn != nil {
		return m.fn(ctx, job)
	}
	return domain.DeliveryJobResult{StatusCode: 200, ResponseBody: "ok"}, nil
}

type mockDeliveryLogger struct {
	logs []*domain.DeliveryLog
}

func (m *mockDeliveryLogger) LogDelivery(ctx context.Context, log *domain.DeliveryLog) error {
	m.logs = append(m.logs, log)
	return nil
}

type failingLogger struct{}

func (f *failingLogger) LogDelivery(ctx context.Context, log *domain.DeliveryLog) error {
	return errors.New("logger failed")
}

func TestNewWorker_NilHandler(t *testing.T) {
	_, err := NewWorker(nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestHandleDelivery_Success(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{StatusCode: 200, ResponseBody: "ok"}, nil
		},
	}
	logger := &mockDeliveryLogger{}
	w := &Worker{handler: handler.handle, deliveryLogger: logger}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		SecretKey:    "secret",
		AttemptCount: 0,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != rabbitmq.Ack {
		t.Errorf("expected Ack, got %v", action)
	}
	if len(logger.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logger.logs))
	}
	if logger.logs[0].Status != "success" {
		t.Errorf("expected status=success, got %s", logger.logs[0].Status)
	}
}

func TestHandleDelivery_RetryableError(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{StatusCode: 500, ResponseBody: "error"}, errors.New("timeout")
		},
	}
	logger := &mockDeliveryLogger{}
	w := &Worker{handler: handler.handle, deliveryLogger: logger}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		AttemptCount: 0,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if action != rabbitmq.NackRequeue {
		t.Errorf("expected NackRequeue, got %v", action)
	}
	if len(logger.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logger.logs))
	}
	if logger.logs[0].Status != "attempt_failed" {
		t.Errorf("expected status=attempt_failed, got %s", logger.logs[0].Status)
	}
}

func TestHandleDelivery_ExhaustedAttempts(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{StatusCode: 500, ResponseBody: "error"}, errors.New("permanent failure")
		},
	}
	logger := &mockDeliveryLogger{}
	w := &Worker{handler: handler.handle, deliveryLogger: logger}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		AttemptCount: domain.MaxDeliveryAttempts - 1,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if action != rabbitmq.NackDiscard {
		t.Errorf("expected NackDiscard, got %v", action)
	}
	if len(logger.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logger.logs))
	}
	if logger.logs[0].Status != "failed" {
		t.Errorf("expected status=failed, got %s", logger.logs[0].Status)
	}
}

func TestHandleDelivery_NilDeliveryLogger(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{}, errors.New("error")
		},
	}
	w := &Worker{handler: handler.handle, deliveryLogger: nil}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		AttemptCount: 0,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if action != rabbitmq.NackRequeue {
		t.Errorf("expected NackRequeue, got %v", action)
	}
}

func TestHandleDelivery_LogDeliveryError(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{StatusCode: 200}, nil
		},
	}
	logger := &failingLogger{}
	w := &Worker{handler: handler.handle, deliveryLogger: logger}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		AttemptCount: 0,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != rabbitmq.Ack {
		t.Errorf("expected Ack, got %v", action)
	}
}

func TestHandleDelivery_AttemptCountIncremented(t *testing.T) {
	handler := &mockDeliveryHandler{
		fn: func(ctx context.Context, job domain.DeliveryJob) (domain.DeliveryJobResult, error) {
			return domain.DeliveryJobResult{StatusCode: 502}, errors.New("bad gateway")
		},
	}
	logger := &mockDeliveryLogger{}
	w := &Worker{handler: handler.handle, deliveryLogger: logger}

	job := domain.DeliveryJob{
		EventID:      "evt-1",
		EventType:    "payment.completed",
		TargetURL:    "https://example.com",
		AttemptCount: 2,
	}

	action, err := w.HandleDelivery(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if action != rabbitmq.NackRequeue {
		t.Errorf("expected NackRequeue, got %v", action)
	}
	if logger.logs[0].AttemptNumber != 3 {
		t.Errorf("expected AttemptNumber=3, got %d", logger.logs[0].AttemptNumber)
	}
	if logger.logs[0].StatusCode != 502 {
		t.Errorf("expected StatusCode=502, got %d", logger.logs[0].StatusCode)
	}
}
