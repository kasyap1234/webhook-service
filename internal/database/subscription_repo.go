package database

import (
	"context"

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
	return []domain.Subscription{}, nil
}
