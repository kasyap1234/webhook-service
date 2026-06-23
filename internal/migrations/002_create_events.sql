-- +goose Up
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_tenant_type ON events (tenant_id, event_type);
CREATE INDEX idx_events_received_at ON events (received_at);

-- +goose Down
DROP TABLE IF EXISTS events;
