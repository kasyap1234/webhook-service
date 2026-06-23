package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

const eventColumns = "id, event_type, tenant_id, payload"

type EventStore struct {
	pool *pgxpool.Pool
}

func NewEventStore(pool *pgxpool.Pool) *EventStore {
	return &EventStore{pool: pool}
}

func (s *EventStore) CreateEvent(ctx context.Context, event *domain.WebhookEvent) error {
	query := `INSERT INTO events (id, event_type, tenant_id, payload) VALUES ($1, $2, $3, $4)`
	_, err := s.pool.Exec(ctx, query, event.ID, event.EventType, event.TenantID, event.Payload)
	return err
}

func (s *EventStore) GetEvents(ctx context.Context, tenantID string) ([]domain.WebhookEvent, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE tenant_id = $1`
	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.WebhookEvent])
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *EventStore) DeleteEvent(ctx context.Context, eventID string) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, eventID)
	return err
}

func (s *EventStore) UpdateEvent(ctx context.Context, event *domain.WebhookEvent) error {
	query := `UPDATE events SET event_type = $1, payload = $2 WHERE id = $3`
	_, err := s.pool.Exec(ctx, query, event.EventType, event.Payload, event.ID)
	return err
}
