package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/ingestion"
	"github.com/kasyap1234/webhook-service/internal/subscription"
)

type Handler struct {
	SubscriptionHandler *SubscriptionHandler
	IngestionHandler    *IngestionHandler
}

func NewHandler(subscriptionService *subscription.SubscriptionService, ingestionService *ingestion.IngestionService) *Handler {
	return &Handler{
		SubscriptionHandler: NewSubscriptionHandler(subscriptionService),
		IngestionHandler:    NewIngestionHandler(ingestionService),
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
