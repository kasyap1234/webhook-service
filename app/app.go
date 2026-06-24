package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const httpClientTimeout = 30 * time.Second

type App struct {
	Router           *gin.Engine
	Config           *config.Config
	Handler          *handlers.Handler
	IngestionService *ingestion.IngestionService
	Worker           *queue.Worker
	Broker           *queue.Broker
	rabbitConn       *rabbitmq.Conn
	pool             *pgxpool.Pool
	idempotency      *ingestion.IdempotencyStore
}

func NewApp() *App {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbURL := cfg.DatabaseURL

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
	eventStore := database.NewEventStore(pool)
	deliveryLogsStore := database.NewDeliveryLogsStore(pool)

	// --- Services ---
	subscriptionService := subscription.NewSubscriptionService(store)
	idempotencyStore := ingestion.NewIdempotencyStore(24 * time.Hour)
	ingestionService := ingestion.NewIngestionService(store, eventStore, broker, idempotencyStore)

	// --- Worker with webhook delivery handler ---
	httpClient := &http.Client{Timeout: httpClientTimeout}
	deliverer := delivery.NewHTTPDeliverer(httpClient)
	worker, err := queue.NewWorker(rabbitConn, deliverer.Deliver, deliveryLogsStore)
	if err != nil {
		log.Fatalf("failed to create delivery worker: %v", err)
	}

	// --- HTTP handlers ---
	handler := handlers.NewHandler(subscriptionService, ingestionService)

	return &App{
		Config:           cfg,
		Router:           gin.Default(),
		Handler:          handler,
		IngestionService: ingestionService,
		Worker:           worker,
		Broker:           broker,
		rabbitConn:       rabbitConn,
		pool:             pool,
		idempotency:      idempotencyStore,
	}
}

func (a *App) SetupRoutes() {
	p := ginprometheus.NewPrometheus("gin")
	p.Use(a.Router)

	a.Router.GET("/health", a.Handler.Health)

	a.Router.POST("/subscriptions/activate", a.Handler.SubscriptionHandler.ActivateSubscription)
	a.Router.POST("/subscriptions/deactivate", a.Handler.SubscriptionHandler.DeactivateSubscription)
	a.Router.POST("/events", a.Handler.IngestionHandler.IngestEvent)
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
	if err := a.Router.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// Close cleanly shuts down all resources.
func (a *App) Close() {
	a.Worker.Close()
	a.idempotency.Close()
	if err := a.Broker.Close(); err != nil {
		log.Printf("error closing broker: %v", err)
	}
	a.rabbitConn.Close()
	a.pool.Close()
}
