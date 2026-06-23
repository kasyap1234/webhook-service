package handlers

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	SubscriptionHandler *SubscriptionHandler
	IngestionHandler    *IngestionHandler
}

func NewHandler(subscriptionService SubscriptionService, ingestionService IngestionServiceInterface) *Handler {
	return &Handler{
		SubscriptionHandler: NewSubscriptionHandler(subscriptionService),
		IngestionHandler:    NewIngestionHandler(ingestionService),
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
