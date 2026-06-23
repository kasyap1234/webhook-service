package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/subscription"
)

type mockSubscriptionService struct {
	activateFn   func(ctx context.Context, tenantID, eventType, targetURL string) (string, error)
	deactivateFn func(ctx context.Context, tenantID, subscriptionID string) error
}

func (m *mockSubscriptionService) ActivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) (string, error) {
	if m.activateFn != nil {
		return m.activateFn(ctx, tenantID, eventType, targetURL)
	}
	return "", nil
}

func (m *mockSubscriptionService) DeactivateSubscription(ctx context.Context, tenantID, subscriptionID string) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, tenantID, subscriptionID)
	}
	return nil
}

func setupRouter(handler *SubscriptionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subscriptions/activate", handler.ActivateSubscription)
	router.POST("/subscriptions/deactivate", handler.DeactivateSubscription)
	return router
}

func TestActivateSubscription_Success(t *testing.T) {
	mock := &mockSubscriptionService{
		activateFn: func(ctx context.Context, tenantID, eventType, targetURL string) (string, error) {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenantID=tenant-1, got %s", tenantID)
			}
			if eventType != "payment.completed" {
				t.Errorf("expected eventType=payment.completed, got %s", eventType)
			}
			if targetURL != "https://example.com/webhook" {
				t.Errorf("expected targetURL=https://example.com/webhook, got %s", targetURL)
			}
			return "abc123secret", nil
		},
	}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id":  "tenant-1",
		"event_type": "payment.completed",
		"target_url": "https://example.com/webhook",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["secretKey"] != "abc123secret" {
		t.Errorf("expected secretKey=abc123secret, got %s", resp["secretKey"])
	}
	if resp["message"] != "subscription activated" {
		t.Errorf("expected message=subscription activated, got %s", resp["message"])
	}
}

func TestActivateSubscription_BadRequest(t *testing.T) {
	mock := &mockSubscriptionService{}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id": "tenant-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestActivateSubscription_ServiceError(t *testing.T) {
	mock := &mockSubscriptionService{
		activateFn: func(ctx context.Context, tenantID, eventType, targetURL string) (string, error) {
			return "", errors.New("database connection failed")
		},
	}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id":  "tenant-1",
		"event_type": "payment.completed",
		"target_url": "https://example.com/webhook",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestDeactivateSubscription_Success(t *testing.T) {
	mock := &mockSubscriptionService{
		deactivateFn: func(ctx context.Context, tenantID, subscriptionID string) error {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenantID=tenant-1, got %s", tenantID)
			}
			if subscriptionID != "sub-123" {
				t.Errorf("expected subscriptionID=sub-123, got %s", subscriptionID)
			}
			return nil
		},
	}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id":       "tenant-1",
		"subscription_id": "sub-123",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "subscription deactivated" {
		t.Errorf("expected message=subscription deactivated, got %s", resp["message"])
	}
}

func TestDeactivateSubscription_NotFound(t *testing.T) {
	mock := &mockSubscriptionService{
		deactivateFn: func(ctx context.Context, tenantID, subscriptionID string) error {
			return subscription.ErrSubscriptionNotFound
		},
	}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id":       "tenant-1",
		"subscription_id": "sub-123",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeactivateSubscription_InternalError(t *testing.T) {
	mock := &mockSubscriptionService{
		deactivateFn: func(ctx context.Context, tenantID, subscriptionID string) error {
			return errors.New("database timeout")
		},
	}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id":       "tenant-1",
		"subscription_id": "sub-123",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestDeactivateSubscription_BadRequest(t *testing.T) {
	mock := &mockSubscriptionService{}
	handler := NewSubscriptionHandler(mock)
	router := setupRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"tenant_id": "tenant-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
