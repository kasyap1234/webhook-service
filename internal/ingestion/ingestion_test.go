package ingestion

import (
	"context"
	"errors"
	"testing"

	"github.com/kasyap1234/webhook-service/internal/domain"
)

type mockSubscriptionFinder struct {
	getFn func(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error)
}

func (m *mockSubscriptionFinder) GetActiveSubscriptions(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
	if m.getFn != nil {
		return m.getFn(ctx, tenantID, eventType)
	}
	return nil, nil
}

type mockEventPersister struct {
	createFn func(ctx context.Context, event *domain.WebhookEvent) error
}

func (m *mockEventPersister) CreateEvent(ctx context.Context, event *domain.WebhookEvent) error {
	if m.createFn != nil {
		return m.createFn(ctx, event)
	}
	return nil
}

type mockPublisher struct {
	publishFn func(ctx context.Context, job domain.DeliveryJob) error
	calls     []domain.DeliveryJob
}

func (m *mockPublisher) Publish(ctx context.Context, job domain.DeliveryJob) error {
	m.calls = append(m.calls, job)
	if m.publishFn != nil {
		return m.publishFn(ctx, job)
	}
	return nil
}

func TestIngestEvent_DuplicateEvent(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()
	idempotency.MarkSeen("evt-dup")

	var createCalled bool
	events := &mockEventPersister{
		createFn: func(ctx context.Context, event *domain.WebhookEvent) error {
			createCalled = true
			return nil
		},
	}
	subs := &mockSubscriptionFinder{}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-dup", EventType: "payment.completed", TenantID: "t1"}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createCalled {
		t.Error("expected CreateEvent not to be called for duplicate event")
	}
}

func TestIngestEvent_CreateEventError(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	events := &mockEventPersister{
		createFn: func(ctx context.Context, event *domain.WebhookEvent) error {
			return errors.New("db write failed")
		},
	}
	subs := &mockSubscriptionFinder{}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1"}

	err := svc.IngestEvent(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from CreateEvent")
	}
}

func TestIngestEvent_GetSubscriptionsError(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	events := &mockEventPersister{}
	subs := &mockSubscriptionFinder{
		getFn: func(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
			return nil, errors.New("query failed")
		},
	}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1"}

	err := svc.IngestEvent(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from GetActiveSubscriptions")
	}
}

func TestIngestEvent_NoSubscriptions(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	events := &mockEventPersister{}
	subs := &mockSubscriptionFinder{
		getFn: func(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
			return nil, nil
		},
	}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1"}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(publisher.calls) != 0 {
		t.Errorf("expected 0 publish calls, got %d", len(publisher.calls))
	}
}

func TestIngestEvent_WithSubscriptions(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	events := &mockEventPersister{}
	subs := &mockSubscriptionFinder{
		getFn: func(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{ID: "sub-1", TargetURL: "https://a.example.com", SecretKey: "key1"},
				{ID: "sub-2", TargetURL: "https://b.example.com", SecretKey: "key2"},
			}, nil
		},
	}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1", Payload: "data"}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(publisher.calls) != 2 {
		t.Errorf("expected 2 publish calls, got %d", len(publisher.calls))
	}
	if publisher.calls[0].TargetURL != "https://a.example.com" {
		t.Errorf("expected first job target https://a.example.com, got %s", publisher.calls[0].TargetURL)
	}
	if publisher.calls[1].TargetURL != "https://b.example.com" {
		t.Errorf("expected second job target https://b.example.com, got %s", publisher.calls[1].TargetURL)
	}
}

func TestIngestEvent_PublishError_Continues(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	events := &mockEventPersister{}
	subs := &mockSubscriptionFinder{
		getFn: func(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{ID: "sub-1", TargetURL: "https://a.example.com", SecretKey: "key1"},
				{ID: "sub-2", TargetURL: "https://b.example.com", SecretKey: "key2"},
			}, nil
		},
	}
	publisher := &mockPublisher{
		publishFn: func(ctx context.Context, job domain.DeliveryJob) error {
			if job.TargetURL == "https://a.example.com" {
				return errors.New("queue full")
			}
			return nil
		},
	}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1"}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(publisher.calls) != 2 {
		t.Errorf("expected 2 publish calls (error doesn't stop loop), got %d", len(publisher.calls))
	}
}

type eventPersisterCalls struct {
	events []*domain.WebhookEvent
}

func TestIngestEvent_PersistsEvent(t *testing.T) {
	idempotency := NewIdempotencyStore(24 * 3600e9)
	defer idempotency.Close()

	var persisted []*domain.WebhookEvent
	events := &mockEventPersister{
		createFn: func(ctx context.Context, event *domain.WebhookEvent) error {
			persisted = append(persisted, event)
			return nil
		},
	}
	subs := &mockSubscriptionFinder{}
	publisher := &mockPublisher{}

	svc := NewIngestionService(subs, events, publisher, idempotency)
	event := domain.WebhookEvent{ID: "evt-1", EventType: "payment.completed", TenantID: "t1", Payload: "data"}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(persisted) != 1 {
		t.Fatalf("expected 1 persisted event, got %d", len(persisted))
	}
	if persisted[0].ID != "evt-1" {
		t.Errorf("expected persisted event ID=evt-1, got %s", persisted[0].ID)
	}
}
