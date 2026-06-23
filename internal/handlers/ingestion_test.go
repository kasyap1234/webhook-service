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
	"github.com/kasyap1234/webhook-service/internal/domain"
)

type mockIngestionService struct {
	ingestFn func(ctx context.Context, event domain.WebhookEvent) error
}

func (m *mockIngestionService) IngestEvent(ctx context.Context, event domain.WebhookEvent) error {
	if m.ingestFn != nil {
		return m.ingestFn(ctx, event)
	}
	return nil
}

func setupIngestionRouter(handler *IngestionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/events", handler.IngestEvent)
	return router
}

func TestIngestionHandler_IngestEvent_Success(t *testing.T) {
	mock := &mockIngestionService{
		ingestFn: func(ctx context.Context, event domain.WebhookEvent) error {
			if event.ID != "evt-1" {
				t.Errorf("expected event ID=evt-1, got %s", event.ID)
			}
			if event.EventType != "payment.completed" {
				t.Errorf("expected event_type=payment.completed, got %s", event.EventType)
			}
			return nil
		},
	}
	handler := NewIngestionHandler(mock)
	router := setupIngestionRouter(handler)

	body, _ := json.Marshal(map[string]any{
		"id":         "evt-1",
		"event_type": "payment.completed",
		"tenant_id":  "tenant-1",
		"payload":    map[string]string{"amount": "100"},
	})

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "event accepted" {
		t.Errorf("expected message=event accepted, got %s", resp["message"])
	}
}

func TestIngestionHandler_IngestEvent_BadRequest(t *testing.T) {
	mock := &mockIngestionService{}
	handler := NewIngestionHandler(mock)
	router := setupIngestionRouter(handler)

	body, _ := json.Marshal(map[string]string{
		"id": "evt-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestIngestionHandler_IngestEvent_EmptyBody(t *testing.T) {
	mock := &mockIngestionService{}
	handler := NewIngestionHandler(mock)
	router := setupIngestionRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestIngestionHandler_IngestEvent_ServiceError(t *testing.T) {
	mock := &mockIngestionService{
		ingestFn: func(ctx context.Context, event domain.WebhookEvent) error {
			return errors.New("internal processing error")
		},
	}
	handler := NewIngestionHandler(mock)
	router := setupIngestionRouter(handler)

	body, _ := json.Marshal(map[string]any{
		"id":         "evt-1",
		"event_type": "payment.completed",
		"tenant_id":  "tenant-1",
		"payload":    "data",
	})

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestIngestionHandler_IngestEvent_InvalidJSON(t *testing.T) {
	mock := &mockIngestionService{}
	handler := NewIngestionHandler(mock)
	router := setupIngestionRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
