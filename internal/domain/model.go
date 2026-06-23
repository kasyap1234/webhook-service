// Package domain
package domain

const MaxDeliveryAttempts = 5

type WebhookEvent struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	TenantID  string `json:"tenant_id"`
	Payload   any    `json:"payload"`
}

type Subscription struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	TargetURL string `json:"target_url"`
	EventType string `json:"event_type"`

	SecretKey string `json:"secret_key"`
	IsActive  bool   `json:"is_active"`
}

type DeliveryJob struct {
	EventID        string `json:"event_id"`
	TenantID       string `json:"tenant_id"`
	SubscriptionID string `json:"subscription_id"`

	EventType    string `json:"event_type"`
	TargetURL    string `json:"target_url"`
	SecretKey    string `json:"secret_key"`
	Payload      any    `json:"payload"`
	AttemptCount int    `json:"attempt_count"`
}
type DeliveryJobResult struct {
	StatusCode   int    `json:"status_code"`
	ResponseBody string `json:"response_body"`
	ErrorMessage string `json:"error_message"`
	DurationMs   int    `json:"duration_ms"`
}

type DeliveryLog struct {
	ID             string `json:"id"`
	EventID        string `json:"event_id"`
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	EventType      string `json:"event_type"`
	TargetURL      string `json:"target_url"`
	AttemptNumber  int    `json:"attempt_number"`
	StatusCode     int    `json:"status_code"`
	Status         string `json:"status"`
	ResponseBody   string `json:"response_body"`
	ErrorMessage   string `json:"error_message"`
	DurationMs     int    `json:"duration_ms"`
	CreatedAt      string `json:"created_at"`
}
