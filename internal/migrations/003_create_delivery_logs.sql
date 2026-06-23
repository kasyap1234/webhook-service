-- +goose Up
CREATE TABLE delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id TEXT NOT NULL REFERENCES events(id),
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    attempt_number INT NOT NULL,
    status_code INT,
    response_body TEXT,
    error_message TEXT,
    delivered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    duration_ms INT
);

CREATE INDEX idx_delivery_logs_event_id ON delivery_logs (event_id);
CREATE INDEX idx_delivery_logs_tenant_id ON delivery_logs (tenant_id);
CREATE INDEX idx_delivery_logs_delivered_at ON delivery_logs (delivered_at);
CREATE INDEX idx_delivery_logs_status ON delivery_logs (status_code);

-- +goose Down
DROP TABLE IF EXISTS delivery_logs;
