package subscription

import (
	"context"
	"errors"
	"testing"

	"github.com/kasyap1234/webhook-service/internal/domain"
)

type mockRepo struct {
	subscribeFn       func(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error
	getSubscriptionFn func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error)
	unsubscribeFn     func(ctx context.Context, tenantID, subscriptionID string) error
}

func (m *mockRepo) Subscribe(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error {
	if m.subscribeFn != nil {
		return m.subscribeFn(ctx, tenantID, eventType, targetURL, secretKey)
	}
	return nil
}

func (m *mockRepo) GetSubscription(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
	if m.getSubscriptionFn != nil {
		return m.getSubscriptionFn(ctx, tenantID, subscriptionID)
	}
	return nil, nil
}

func (m *mockRepo) Unsubscribe(ctx context.Context, tenantID, subscriptionID string) error {
	if m.unsubscribeFn != nil {
		return m.unsubscribeFn(ctx, tenantID, subscriptionID)
	}
	return nil
}

func TestActivateSubscription_Success(t *testing.T) {
	repo := &mockRepo{
		subscribeFn: func(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenantID=tenant-1, got %s", tenantID)
			}
			if eventType != "payment.completed" {
				t.Errorf("expected eventType=payment.completed, got %s", eventType)
			}
			if secretKey == "" {
				t.Error("expected non-empty secretKey")
			}
			return nil
		},
	}
	svc := NewSubscriptionService(repo)

	key, err := svc.ActivateSubscription(context.Background(), "tenant-1", "payment.completed", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == "" {
		t.Error("expected non-empty secret key")
	}
}

func TestActivateSubscription_RepoError(t *testing.T) {
	repo := &mockRepo{
		subscribeFn: func(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error {
			return errors.New("database error")
		},
	}
	svc := NewSubscriptionService(repo)

	_, err := svc.ActivateSubscription(context.Background(), "tenant-1", "payment.completed", "https://example.com")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestDeactivateSubscription_Success(t *testing.T) {
	repo := &mockRepo{
		getSubscriptionFn: func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
			return &domain.Subscription{
				ID:       subscriptionID,
				TenantID: tenantID,
				IsActive: true,
			}, nil
		},
		unsubscribeFn: func(ctx context.Context, tenantID, subscriptionID string) error {
			return nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeactivateSubscription(context.Background(), "tenant-1", "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeactivateSubscription_NotFound_NilSubscription(t *testing.T) {
	repo := &mockRepo{
		getSubscriptionFn: func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
			return nil, nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeactivateSubscription(context.Background(), "tenant-1", "sub-123")
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestDeactivateSubscription_NotFound_Inactive(t *testing.T) {
	repo := &mockRepo{
		getSubscriptionFn: func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
			return &domain.Subscription{
				ID:       subscriptionID,
				TenantID: tenantID,
				IsActive: false,
			}, nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeactivateSubscription(context.Background(), "tenant-1", "sub-123")
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestDeactivateSubscription_GetSubscriptionError(t *testing.T) {
	repo := &mockRepo{
		getSubscriptionFn: func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
			return nil, errors.New("database error")
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeactivateSubscription(context.Background(), "tenant-1", "sub-123")
	if err == nil {
		t.Fatal("expected error from GetSubscription")
	}
}

func TestDeactivateSubscription_UnsubscribeError(t *testing.T) {
	repo := &mockRepo{
		getSubscriptionFn: func(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
			return &domain.Subscription{
				ID:       subscriptionID,
				TenantID: tenantID,
				IsActive: true,
			}, nil
		},
		unsubscribeFn: func(ctx context.Context, tenantID, subscriptionID string) error {
			return errors.New("delete failed")
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeactivateSubscription(context.Background(), "tenant-1", "sub-123")
	if err == nil {
		t.Fatal("expected error from Unsubscribe")
	}
}
