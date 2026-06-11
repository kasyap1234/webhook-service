package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/queue"
)

type App struct {
	Repo   *database.SubscriptionRepo
	broker *queue.Broker
}

func NewApp(pool *pgxpool.Pool, broker *queue.Broker) *App {
	return &App{
		Repo:   database.NewSubscriptionRepo(pool),
		broker: broker,
	}
}

