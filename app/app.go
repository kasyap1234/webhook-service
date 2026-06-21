package app

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/config"
	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/queue"
)

type App struct {
	Repo   *database.SubscriptionRepo
	Router *gin.Engine
	Config *config.Config
	Broker *queue.Broker
}

func NewApp() *App {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbURL)
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	connURL := cfg.RabbitMQURL

	broker, err := queue.NewBroker(connURL)
	if err != nil {
		panic(err)
	}

	return &App{
		Repo:   database.NewSubscriptionRepo(pool),
		Broker: broker,
		Config: cfg,
		Router: gin.Default()}
}

func (a *App) Run() {
	a.Router.Run()
}
