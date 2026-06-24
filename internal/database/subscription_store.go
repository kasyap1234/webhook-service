// Package database
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

const subscriptionColumns = "id, tenant_id, event_type, target_url, secret_key, is_active"

type SubscriptionStore struct {
	pool *pgxpool.Pool
}

func NewSubscriptionStore(pool *pgxpool.Pool) *SubscriptionStore {
	return &SubscriptionStore{
		pool: pool,
	}
}

func (r *SubscriptionStore) GetActiveSubscriptions(ctx context.Context, tenantID, eventType string) ([]domain.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE tenant_id = $1 AND event_type = $2 AND is_active = true`
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

func (r *SubscriptionStore) Subscribe(ctx context.Context, tenantID, eventType, targetURL, secretKey string) error {
	updateQuery := `
		UPDATE subscriptions
		SET is_active = true, secret_key = $4
		WHERE tenant_id = $1 AND event_type = $2 AND target_url = $3
	`
	result, err := r.pool.Exec(ctx, updateQuery, tenantID, eventType, targetURL, secretKey)
	if err != nil {
		return err
	}
	if result.RowsAffected() > 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO subscriptions (tenant_id, event_type, target_url, is_active, secret_key)
		VALUES ($1, $2, $3, true, $4)
	`
	_, err = r.pool.Exec(ctx, insertQuery, tenantID, eventType, targetURL, secretKey)
	return err
}

func (r *SubscriptionStore) Unsubscribe(ctx context.Context, tenantID, subscriptionID string) error {
	query := `UPDATE subscriptions SET is_active = false WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, subscriptionID)
	return err
}

func (r *SubscriptionStore) GetSubscriptions(ctx context.Context, tenantID string) ([]domain.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE tenant_id = $1`
	rows, err := r.pool.Query(ctx, query, tenantID)
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

func (r *SubscriptionStore) GetAllSubscriptions(ctx context.Context) ([]domain.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions`
	rows, err := r.pool.Query(ctx, query)
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

func (r *SubscriptionStore) GetSubscription(ctx context.Context, tenantID, subscriptionID string) (*domain.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE tenant_id = $1 AND id = $2`
	rows, err := r.pool.Query(ctx, query, tenantID, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Subscription])
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	return &subscriptions[0], nil
}
