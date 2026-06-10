// Package domain 
package domain 


type WebhookEvent struct {
	ID string `json:"id"`
	EventType string `json:"event_type"`
	TenantID string `json:"tenant_id"`
	Payload any  `json:"payload"`
}


type Subscription struct {
	ID string `json:"id"`
	TenantID string `json:"tenant_id"`
	TargetURL string `json:"target_url"`
	SecretKey string `json:"secret_key"`
	IsActive bool `json:"is_active"`
}

type DeliveryJob struct {
	EventID string `json:"event_id"`
	EventType string `json:"event_type"`
	TargetURL string `json:"target_url"`
	SecretKey string `json"secret_key"`
	Payload any `json:"payload"`
	AttemptCount int `json:"attempt_count"`
}



type SubscriptionRepository interface {
	GetActiveSubscriptions(tenantID string, eventType string)([]Subscription,error)
}