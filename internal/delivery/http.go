package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kasyap1234/webhook-service/internal/domain"
	"github.com/kasyap1234/webhook-service/internal/security"
)

type HTTPDeliverer struct {
	client *http.Client
}

func NewHTTPDeliverer(client *http.Client) *HTTPDeliverer {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPDeliverer{client: client}
}

func (d *HTTPDeliverer) Deliver(ctx context.Context, job domain.DeliveryJob) error {
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	signature, err := security.GenerateSignature(payloadBytes, job.SecretKey)
	if err != nil {
		return fmt.Errorf("sign payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.TargetURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("create delivery request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", job.EventType)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("send delivery request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("target returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
