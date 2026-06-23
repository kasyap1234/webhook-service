-- +goose Up
CREATE TABLE delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id TEXT NOT NULL REFERENCES events(id),
    subscription_id UUID REFERENCES subscriptions(id),
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    attempt_number INT NOT NULL,
    status_code INT,
    status TEXT NOT NULL DEFAULT 'pending',
    response_body TEXT,
    error_message TEXT,
    duration_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_delivery_logs_event_id ON delivery_logs (event_id);
CREATE INDEX idx_delivery_logs_tenant_id ON delivery_logs (tenant_id);
CREATE INDEX idx_delivery_logs_subscription_id ON delivery_logs (subscription_id);
CREATE INDEX idx_delivery_logs_status ON delivery_logs (status);
CREATE INDEX idx_delivery_logs_created_at ON delivery_logs (created_at);

-- +goose Down
DROP TABLE IF EXISTS delivery_logs;
