-- +goose Up
CREATE TABLE failed_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    secret_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    total_attempts INT NOT NULL,
    last_error TEXT,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reprocessed BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_failed_deliveries_tenant ON failed_deliveries (tenant_id);
CREATE INDEX idx_failed_deliveries_reprocessed ON failed_deliveries (reprocessed);

-- +goose Down
DROP TABLE IF EXISTS failed_deliveries;
