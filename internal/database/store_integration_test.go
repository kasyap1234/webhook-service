//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/testutil"
)

func TestSubscriptionStore_Subscribe_New(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	err := store.Subscribe(ctx, "tenant-1", "payment.completed", "https://example.com/webhook", "secret-key-1")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	subs, err := store.GetActiveSubscriptions(ctx, "tenant-1", "payment.completed")
	if err != nil {
		t.Fatalf("GetActiveSubscriptions failed: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].TargetURL != "https://example.com/webhook" {
		t.Errorf("expected target URL https://example.com/webhook, got %s", subs[0].TargetURL)
	}
	if subs[0].SecretKey != "secret-key-1" {
		t.Errorf("expected secret key secret-key-1, got %s", subs[0].SecretKey)
	}
	if !subs[0].IsActive {
		t.Error("expected subscription to be active")
	}
}

func TestSubscriptionStore_Subscribe_Existing(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	err := store.Subscribe(ctx, "tenant-2", "order.created", "https://example.com/orders", "old-secret")
	if err != nil {
		t.Fatalf("first Subscribe failed: %v", err)
	}

	err = store.Subscribe(ctx, "tenant-2", "order.created", "https://example.com/orders", "new-secret")
	if err != nil {
		t.Fatalf("second Subscribe failed: %v", err)
	}

	subs, err := store.GetActiveSubscriptions(ctx, "tenant-2", "order.created")
	if err != nil {
		t.Fatalf("GetActiveSubscriptions failed: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription (upsert), got %d", len(subs))
	}
	if subs[0].SecretKey != "new-secret" {
		t.Errorf("expected secret key new-secret, got %s", subs[0].SecretKey)
	}
}

func TestSubscriptionStore_Unsubscribe(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	err := store.Subscribe(ctx, "tenant-3", "payment.failed", "https://example.com/fail", "secret")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	subs, err := store.GetActiveSubscriptions(ctx, "tenant-3", "payment.failed")
	if err != nil {
		t.Fatalf("GetActiveSubscriptions failed: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}

	err = store.Unsubscribe(ctx, "tenant-3", subs[0].ID)
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	subs, err = store.GetActiveSubscriptions(ctx, "tenant-3", "payment.failed")
	if err != nil {
		t.Fatalf("GetActiveSubscriptions failed: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected 0 active subscriptions after unsubscribe, got %d", len(subs))
	}
}

func TestSubscriptionStore_GetSubscription_Found(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	err := store.Subscribe(ctx, "tenant-4", "user.created", "https://example.com/users", "secret")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	subs, err := store.GetActiveSubscriptions(ctx, "tenant-4", "user.created")
	if err != nil {
		t.Fatalf("GetActiveSubscriptions failed: %v", err)
	}

	sub, err := store.GetSubscription(ctx, "tenant-4", subs[0].ID)
	if err != nil {
		t.Fatalf("GetSubscription failed: %v", err)
	}
	if sub == nil {
		t.Fatal("expected subscription to be found")
	}
	if sub.EventType != "user.created" {
		t.Errorf("expected event_type=user.created, got %s", sub.EventType)
	}
}

func TestSubscriptionStore_GetSubscription_NotFound(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	sub, err := store.GetSubscription(ctx, "tenant-5", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetSubscription failed: %v", err)
	}
	if sub != nil {
		t.Error("expected nil subscription for non-existent ID")
	}
}

func TestSubscriptionStore_GetAllSubscriptions(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	store.Subscribe(ctx, "tenant-A", "evt-1", "https://a.example.com", "secret-a")
	store.Subscribe(ctx, "tenant-B", "evt-2", "https://b.example.com", "secret-b")

	subs, err := store.GetAllSubscriptions(ctx)
	if err != nil {
		t.Fatalf("GetAllSubscriptions failed: %v", err)
	}
	if len(subs) < 2 {
		t.Errorf("expected at least 2 subscriptions, got %d", len(subs))
	}
}

func TestSubscriptionStore_GetSubscriptions_ByTenant(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	store := database.NewSubscriptionStore(deps.Pool)

	store.Subscribe(ctx, "tenant-X", "evt-1", "https://x1.example.com", "secret-x1")
	store.Subscribe(ctx, "tenant-X", "evt-2", "https://x2.example.com", "secret-x2")
	store.Subscribe(ctx, "tenant-Y", "evt-1", "https://y1.example.com", "secret-y1")

	subs, err := store.GetSubscriptions(ctx, "tenant-X")
	if err != nil {
		t.Fatalf("GetSubscriptions failed: %v", err)
	}
	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions for tenant-X, got %d", len(subs))
	}
}

func TestEventStore_CreateAndGet(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)

	event := &domain.WebhookEvent{
		ID:        "evt-100",
		EventType: "payment.completed",
		TenantID:  "tenant-evt",
		Payload:   map[string]any{"amount": 50},
	}

	err := eventStore.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	events, err := eventStore.GetEvents(ctx, "tenant-evt")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "evt-100" {
		t.Errorf("expected event ID=evt-100, got %s", events[0].ID)
	}
}

func TestEventStore_DeleteEvent(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)

	event := &domain.WebhookEvent{
		ID:        "evt-del",
		EventType: "order.cancelled",
		TenantID:  "tenant-del",
		Payload:   "data",
	}

	eventStore.CreateEvent(ctx, event)
	err := eventStore.DeleteEvent(ctx, "evt-del")
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	events, _ := eventStore.GetEvents(ctx, "tenant-del")
	if len(events) != 0 {
		t.Errorf("expected 0 events after delete, got %d", len(events))
	}
}

func TestEventStore_UpdateEvent(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)

	event := &domain.WebhookEvent{
		ID:        "evt-upd",
		EventType: "payment.pending",
		TenantID:  "tenant-upd",
		Payload:   "old",
	}

	eventStore.CreateEvent(ctx, event)

	event.EventType = "payment.completed"
	event.Payload = "new"
	err := eventStore.UpdateEvent(ctx, event)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	events, _ := eventStore.GetEvents(ctx, "tenant-upd")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "payment.completed" {
		t.Errorf("expected updated event_type=payment.completed, got %s", events[0].EventType)
	}
}

func TestEventStore_GetEvents_Empty(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)

	events, err := eventStore.GetEvents(ctx, "non-existent-tenant")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for non-existent tenant, got %d", len(events))
	}
}

func TestDeliveryLogsStore_LogDelivery(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)
	deliveryStore := database.NewDeliveryLogsStore(deps.Pool)

	eventStore.CreateEvent(ctx, &domain.WebhookEvent{
		ID:        "evt-log",
		EventType: "test.event",
		TenantID:  "tenant-log",
		Payload:   "data",
	})

	logEntry := &domain.DeliveryLog{
		EventID:        "evt-log",
		TenantID:       "tenant-log",
		EventType:      "test.event",
		TargetURL:      "https://example.com",
		AttemptNumber:  1,
		StatusCode:     200,
		Status:         "success",
		ResponseBody:   "ok",
		DurationMs:     150,
	}

	err := deliveryStore.LogDelivery(ctx, logEntry)
	if err != nil {
		t.Fatalf("LogDelivery failed: %v", err)
	}

	logs, err := deliveryStore.GetDeliveryLogs(ctx)
	if err != nil {
		t.Fatalf("GetDeliveryLogs failed: %v", err)
	}
	if len(logs) < 1 {
		t.Fatalf("expected at least 1 log, got %d", len(logs))
	}
}

func TestDeliveryLogsStore_GetDeliveryLogsByEvent(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)
	deliveryStore := database.NewDeliveryLogsStore(deps.Pool)

	eventStore.CreateEvent(ctx, &domain.WebhookEvent{
		ID:        "evt-filter",
		EventType: "filter.test",
		TenantID:  "tenant-filter",
		Payload:   "data",
	})

	deliveryStore.LogDelivery(ctx, &domain.DeliveryLog{
		EventID: "evt-filter", TenantID: "tenant-filter", EventType: "filter.test",
		TargetURL: "https://example.com", AttemptNumber: 1, StatusCode: 200, Status: "success", DurationMs: 100,
	})
	deliveryStore.LogDelivery(ctx, &domain.DeliveryLog{
		EventID: "other-evt", TenantID: "tenant-filter", EventType: "other.test",
		TargetURL: "https://example.com", AttemptNumber: 1, StatusCode: 200, Status: "success", DurationMs: 100,
	})

	logs, err := deliveryStore.GetDeliveryLogsByEvent(ctx, "evt-filter")
	if err != nil {
		t.Fatalf("GetDeliveryLogsByEvent failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log for event, got %d", len(logs))
	}
}

func TestDeliveryLogsStore_GetDeliveryLogsByStatus(t *testing.T) {
	testutil.SkipIfShort(t)
	deps := testutil.Setup(t)
	ctx := context.Background()

	eventStore := database.NewEventStore(deps.Pool)
	deliveryStore := database.NewDeliveryLogsStore(deps.Pool)

	eventStore.CreateEvent(ctx, &domain.WebhookEvent{
		ID:        "evt-status",
		EventType: "status.test",
		TenantID:  "tenant-status",
		Payload:   "data",
	})

	deliveryStore.LogDelivery(ctx, &domain.DeliveryLog{
		EventID: "evt-status", TenantID: "tenant-status", EventType: "status.test",
		TargetURL: "https://example.com", AttemptNumber: 1, StatusCode: 200, Status: "success", DurationMs: 100,
	})
	deliveryStore.LogDelivery(ctx, &domain.DeliveryLog{
		EventID: "evt-status", TenantID: "tenant-status", EventType: "status.test",
		TargetURL: "https://example.com", AttemptNumber: 2, StatusCode: 500, Status: "failed", DurationMs: 200,
	})

	logs, err := deliveryStore.GetDeliveryLogsByStatus(ctx, "tenant-status", "failed")
	if err != nil {
		t.Fatalf("GetDeliveryLogsByStatus failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 failed log, got %d", len(logs))
	}
	if logs[0].Status != "failed" {
		t.Errorf("expected status=failed, got %s", logs[0].Status)
	}
}
