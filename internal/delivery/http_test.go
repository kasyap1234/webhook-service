package delivery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/security"
)

func TestHTTPDeliverer_Deliver_Success(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: "test-secret-key",
		Payload:   map[string]string{"amount": "100"},
	}

	result, err := deliverer.Deliver(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if result.ResponseBody != `{"ok":true}` {
		t.Errorf("expected body={\"ok\":true}, got %s", result.ResponseBody)
	}
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %s", receivedHeaders.Get("Content-Type"))
	}
	if receivedHeaders.Get("X-Webhook-Event") != "payment.completed" {
		t.Errorf("expected X-Webhook-Event=payment.completed, got %s", receivedHeaders.Get("X-Webhook-Event"))
	}
	if receivedHeaders.Get("X-Webhook-Signature") == "" {
		t.Error("expected X-Webhook-Signature header to be set")
	}
	if len(receivedBody) == 0 {
		t.Error("expected non-empty request body")
	}
}

func TestHTTPDeliverer_Deliver_SignatureValid(t *testing.T) {
	var receivedSignature string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Webhook-Signature")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	secretKey := "my-webhook-secret"
	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: secretKey,
		Payload:   map[string]string{"key": "value"},
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !security.VerifySignature(receivedBody, secretKey, receivedSignature) {
		t.Error("signature verification failed - signature is not valid for the given payload and secret")
	}
}

func TestHTTPDeliverer_Deliver_Non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: "test-secret",
		Payload:   map[string]string{"key": "value"},
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestHTTPDeliverer_Deliver_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: "test-secret",
		Payload:   map[string]string{"key": "value"},
	}

	_, err := deliverer.Deliver(ctx, job)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestHTTPDeliverer_Deliver_ConnectionRefused(t *testing.T) {
	deliverer := NewHTTPDeliverer(&http.Client{})
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: "http://127.0.0.1:1",
		SecretKey: "test-secret",
		Payload:   map[string]string{"key": "value"},
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestHTTPDeliverer_Deliver_InvalidPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: "test-secret",
		Payload:   make(chan int),
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for non-marshalable payload")
	}
}

func TestNewHTTPDeliverer_NilClient(t *testing.T) {
	deliverer := NewHTTPDeliverer(nil)
	if deliverer.client != http.DefaultClient {
		t.Error("expected nil client to default to http.DefaultClient")
	}
}

func TestHTTPDeliverer_Deliver_EventHeader(t *testing.T) {
	var receivedEventHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEventHeader = r.Header.Get("X-Webhook-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	deliverer := NewHTTPDeliverer(server.Client())
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "order.created",
		TargetURL: server.URL,
		SecretKey: "test-secret",
		Payload:   map[string]string{"key": "value"},
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedEventHeader != "order.created" {
		t.Errorf("expected X-Webhook-Event=order.created, got %s", receivedEventHeader)
	}
}

func TestHTTPDeliverer_Deliver_PayloadMarshaling(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	deliverer := NewHTTPDeliverer(server.Client())
	payload := map[string]any{
		"amount": 99.99,
		"items":  []string{"a", "b"},
	}
	job := domain.DeliveryJob{
		EventID:   "evt-1",
		EventType: "payment.completed",
		TargetURL: server.URL,
		SecretKey: "test-secret",
		Payload:   payload,
	}

	_, err := deliverer.Deliver(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(receivedBody, &decoded); err != nil {
		t.Fatalf("failed to unmarshal received body: %v", err)
	}
	if decoded["amount"] != 99.99 {
		t.Errorf("expected amount=99.99, got %v", decoded["amount"])
	}
}
