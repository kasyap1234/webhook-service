package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kasyap1234/webhook-service/internal/config"
	"github.com/kasyap1234/webhook-service/internal/database"
	"github.com/kasyap1234/webhook-service/internal/delivery"
	"github.com/kasyap1234/webhook-service/internal/handlers"
	"github.com/kasyap1234/webhook-service/internal/ingestion"
	"github.com/kasyap1234/webhook-service/internal/queue"
	"github.com/kasyap1234/webhook-service/internal/subscription"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type App struct {
	Router           *gin.Engine
	Config           *config.Config
	Handler          *handlers.Handler
	IngestionService *ingestion.IngestionService
	Worker           *queue.Worker
	Broker           *queue.Broker
	rabbitConn       *rabbitmq.Conn
	pool             *pgxpool.Pool
}

func NewApp() *App {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/webhooks?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}

	// Create shared RabbitMQ connection
	rabbitConn, err := rabbitmq.NewConn(cfg.RabbitMQURL, rabbitmq.WithConnectionOptionsLogging)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	// --- Dependencies ---
	broker, err := queue.NewBroker(rabbitConn)
	if err != nil {
		log.Fatalf("failed to create message broker: %v", err)
	}

	store := database.NewSubscriptionStore(pool)

	// --- Services ---
	subscriptionService := subscription.NewSubscriptionService(store)
	ingestionService := ingestion.NewIngestionService(store, broker)

	// --- Worker with webhook delivery handler ---
	deliverer := delivery.NewHTTPDeliverer(http.DefaultClient)
	worker, err := queue.NewWorker(rabbitConn, deliverer.Deliver)
	if err != nil {
		log.Fatalf("failed to create delivery worker: %v", err)
	}

	// --- HTTP handlers ---
	handler := handlers.NewHandler(subscriptionService)

	return &App{
		Config:           cfg,
		Router:           gin.Default(),
		Handler:          handler,
		IngestionService: ingestionService,
		Worker:           worker,
		Broker:           broker,
		rabbitConn:       rabbitConn,
		pool:             pool,
	}
}

func (a *App) SetupRoutes() {
	a.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	a.Router.POST("/subscriptions/activate", a.Handler.SubscriptionHandler.ActivateSubscription)
	a.Router.POST("/subscriptions/deactivate", a.Handler.SubscriptionHandler.DeactivateSubscription)
}

func (a *App) Run() {
	a.SetupRoutes()

	// Start the worker in the background
	go func() {
		log.Println("starting webhook delivery worker")
		if err := a.Worker.Start(); err != nil {
			log.Fatalf("worker exited with error: %v", err)
		}
	}()

	addr := fmt.Sprintf(":%d", a.Config.Port)
	log.Printf("starting server on %s", addr)
	a.Router.Run(addr)
}
