//go:build integration

package ingestion_test

import (
	"context"
	"testing"
	"time"

	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/ingestion"
	"github.com/kasyap1234/webhook-service/internal/testutil"
)

func TestIngestEvent_FanOut(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	subStore := database.NewSubscriptionStore(deps.Pool)
	eventStore := database.NewEventStore(deps.Pool)
	idempotency := ingestion.NewIdempotencyStore(24 * time.Hour)
	defer idempotency.Close()

	subStore.Subscribe(ctx, "tenant-fan", "payment.completed", "https://a.example.com", "secret-a")
	subStore.Subscribe(ctx, "tenant-fan", "payment.completed", "https://b.example.com", "secret-b")

	svc := ingestion.NewIngestionService(subStore, eventStore, &noopPublisher{}, idempotency)

	event := domain.WebhookEvent{
		ID:        "evt-fan-1",
		EventType: "payment.completed",
		TenantID:  "tenant-fan",
		Payload:   map[string]any{"amount": 100},
	}

	err := svc.IngestEvent(ctx, event)
	if err != nil {
		t.Fatalf("IngestEvent failed: %v", err)
	}

	events, err := eventStore.GetEvents(ctx, "tenant-fan")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event persisted, got %d", len(events))
	}
	if events[0].ID != "evt-fan-1" {
		t.Errorf("expected event ID=evt-fan-1, got %s", events[0].ID)
	}
}

func TestIngestEvent_DuplicateEvent_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	subStore := database.NewSubscriptionStore(deps.Pool)
	eventStore := database.NewEventStore(deps.Pool)
	idempotency := ingestion.NewIdempotencyStore(24 * time.Hour)
	defer idempotency.Close()

	svc := ingestion.NewIngestionService(subStore, eventStore, &noopPublisher{}, idempotency)

	event := domain.WebhookEvent{
		ID:        "evt-dup-int",
		EventType: "test.duplicate",
		TenantID:  "tenant-dup",
		Payload:   "data",
	}

	err := svc.IngestEvent(ctx, event)
	if err != nil {
		t.Fatalf("first IngestEvent failed: %v", err)
	}

	err = svc.IngestEvent(ctx, event)
	if err != nil {
		t.Fatalf("second IngestEvent failed: %v", err)
	}

	events, _ := eventStore.GetEvents(ctx, "tenant-dup")
	if len(events) != 1 {
		t.Errorf("expected 1 event (dedup), got %d", len(events))
	}
}

func TestIngestEvent_NoSubscribers_Integration(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	subStore := database.NewSubscriptionStore(deps.Pool)
	eventStore := database.NewEventStore(deps.Pool)
	idempotency := ingestion.NewIdempotencyStore(24 * time.Hour)
	defer idempotency.Close()

	publisher := &trackingPublisher{}

	svc := ingestion.NewIngestionService(subStore, eventStore, publisher, idempotency)

	event := domain.WebhookEvent{
		ID:        "evt-nosub",
		EventType: "no.subscribers",
		TenantID:  "tenant-nosub",
		Payload:   "data",
	}

	err := svc.IngestEvent(ctx, event)
	if err != nil {
		t.Fatalf("IngestEvent failed: %v", err)
	}

	if len(publisher.jobs) != 0 {
		t.Errorf("expected 0 publish calls (no subscribers), got %d", len(publisher.jobs))
	}
}

type noopPublisher struct{}

func (n *noopPublisher) Publish(ctx context.Context, job domain.DeliveryJob) error {
	return nil
}

type trackingPublisher struct {
	jobs []domain.DeliveryJob
}

func (t *trackingPublisher) Publish(ctx context.Context, job domain.DeliveryJob) error {
	t.jobs = append(t.jobs, job)
	return nil
}
