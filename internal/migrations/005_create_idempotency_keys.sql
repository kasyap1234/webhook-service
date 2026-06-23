-- +goose Up
CREATE TABLE idempotency_keys (
    event_id TEXT PRIMARY KEY,
    seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_idempotency_keys_seen_at ON idempotency_keys (seen_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
