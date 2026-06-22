# Webhook Service

A Go service that ingests webhook events and delivers them to registered subscriber endpoints via a RabbitMQ queue.

## Architecture

```
Event Source → POST /events → Ingestion Service → RabbitMQ → Worker → HMAC-signed POST → Subscriber
                                     ↓
                               PostgreSQL
                            (subscriptions)
```

**Components:**
- **Ingestion API** — receives events, matches to active subscriptions, publishes delivery jobs
- **Queue Worker** — consumes delivery jobs, signs payloads, delivers via HTTP
- **Subscription API** — manages subscriber registrations (activate/deactivate)

## Setup

### Prerequisites
- Go 1.24+
- PostgreSQL
- RabbitMQ

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `postgres://localhost:5432/webhooks?sslmode=disable` | PostgreSQL connection string |
| `RABBITMQ_URL` | — | RabbitMQ connection URL |

### Run with Docker

```bash
docker-compose up --build
```

### Run locally

```bash
# Start dependencies
docker-compose up -d postgres rabbitmq

# Run the service
go run ./cmd/main.go
```

### Database Setup

Create the `subscriptions` table:

```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    secret_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, event_type, target_url)
);
```

## API

### Health Check
```
GET /health
→ 200 {"status": "ok"}
```

### Activate Subscription
```
POST /subscriptions/activate
Content-Type: application/json

{
  "tenant_id": "tenant-1",
  "event_type": "payment.completed",
  "target_url": "https://example.com/webhook"
}

→ 200 {"message": "subscription activated", "secretKey": "..."}
```

### Deactivate Subscription
```
POST /subscriptions/deactivate
Content-Type: application/json

{
  "tenant_id": "tenant-1",
  "subscription_id": "1"
}

→ 200 {"message": "subscription deactivated"}
→ 404 {"error": "subscription not found or already inactive"}
```

### Ingest Event
```
POST /events
Content-Type: application/json

{
  "id": "evt-001",
  "event_type": "payment.completed",
  "tenant_id": "tenant-1",
  "payload": {"amount": 100, "currency": "USD"}
}

→ 202 {"message": "event accepted"}
```

## Webhook Delivery

The worker signs each outgoing webhook payload with HMAC-SHA256 using the subscriber's secret key. Subscribers should verify:

```go
import "github.com/kasyap1234/webhook-service/internal/security"

valid := security.VerifySignature(body, secretKey, r.Header.Get("X-Webhook-Signature"))
```

**Headers sent:**
- `X-Webhook-Signature` — HMAC-SHA256 hex digest of the JSON payload
- `X-Webhook-Event` — the event type
- `Content-Type: application/json`

## Development

```bash
# Run tests
go test ./...

# Lint
golangci-lint run

# Build
go build -o bin/webhook-service ./cmd/main.go
```
