-- +goose Up
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    secret_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, event_type, target_url)
);

CREATE INDEX idx_subscriptions_active ON subscriptions (tenant_id, event_type) WHERE is_active = true;

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
