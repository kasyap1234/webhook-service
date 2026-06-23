package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

type EventStore interface {
}

type eventStore struct {
	pool *pgxpool.Pool 
}


func(e*eventStore)IngestEvent(ctx context.Context,event *domain.WebhookEvent)(error){
	query :=`INSERT INTO events (id, event_type, tenant_id, payload) VALUES ($1, $2, $3, $4)`
	_, err := e.pool.Exec(ctx, query, event.ID, event.EventType, event.TenantID, event.Payload)
	return err
	
}

func(e*eventStore) GetEvents(ctx context.Context, tenantID string) ([]domain.WebhookEvent, error) {
	query := `SELECT id, event_type, tenant_id, payload FROM events WHERE tenant_id = $1`
	rows, err := e.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.WebhookEvent
	for rows.Next() {
		var event domain.WebhookEvent
		if err := rows.Scan(&event.ID, &event.EventType, &event.TenantID, &event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}



func (e *eventStore) DeleteEvent(ctx context.Context, eventID string) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := e.pool.Exec(ctx, query, eventID)
	return err
}

func (e *eventStore) UpdateEvent(ctx context.Context, event *domain.WebhookEvent) error {
	query := `UPDATE events SET event_type = $1, payload = $2 WHERE id = $3`
	_, err := e.pool.Exec(ctx, query, event.EventType, event.Payload, event.ID)
	return err	
}



