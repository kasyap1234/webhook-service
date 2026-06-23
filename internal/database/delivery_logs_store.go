package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

const deliveryLogColumns = "id, event_id, subscription_id, tenant_id, event_type, target_url, attempt_number, status_code, status, response_body, error_message, duration_ms, created_at"

type DeliveryLogsStore struct {
	pool *pgxpool.Pool
}

func NewDeliveryLogsStore(pool *pgxpool.Pool) *DeliveryLogsStore {
	return &DeliveryLogsStore{pool: pool}
}

func (s *DeliveryLogsStore) LogDelivery(ctx context.Context, log *domain.DeliveryLog) error {
	query := `INSERT INTO delivery_logs (event_id, subscription_id, tenant_id, event_type, target_url, attempt_number, status_code, status, response_body, error_message, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := s.pool.Exec(ctx, query,
		log.EventID, log.SubscriptionID, log.TenantID, log.EventType,
		log.TargetURL, log.AttemptNumber, log.StatusCode, log.Status,
		log.ResponseBody, log.ErrorMessage, log.DurationMs)
	return err
}

func (s *DeliveryLogsStore) GetDeliveryLogs(ctx context.Context) ([]domain.DeliveryLog, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+deliveryLogColumns+" FROM delivery_logs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DeliveryLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *DeliveryLogsStore) GetDeliveryLogByID(ctx context.Context, id string) (*domain.DeliveryLog, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+deliveryLogColumns+" FROM delivery_logs WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DeliveryLog])
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, nil
	}
	return &logs[0], nil
}

func (s *DeliveryLogsStore) GetDeliveryLogsByEvent(ctx context.Context, eventID string) ([]domain.DeliveryLog, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+deliveryLogColumns+" FROM delivery_logs WHERE event_id = $1", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DeliveryLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *DeliveryLogsStore) GetDeliveryLogsByTenant(ctx context.Context, tenantID string) ([]domain.DeliveryLog, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+deliveryLogColumns+" FROM delivery_logs WHERE tenant_id = $1", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DeliveryLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *DeliveryLogsStore) GetDeliveryLogsByStatus(ctx context.Context, tenantID, status string) ([]domain.DeliveryLog, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+deliveryLogColumns+" FROM delivery_logs WHERE tenant_id = $1 AND status = $2", tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DeliveryLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}
