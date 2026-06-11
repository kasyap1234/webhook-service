// Package database 
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{
		pool: pool,
	}
}


func (r *SubscriptionRepo) GetActiveSubscriptions(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
	query := `SELECT * FROM subscriptions WHERE tenant_id = $1 AND event_type = $2 AND is_active = true`
	rows, err := r.pool.Query(ctx, query, tenantID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Subscription])
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}


