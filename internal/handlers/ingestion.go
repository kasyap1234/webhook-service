package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/domain"
)

type IngestionServiceInterface interface {
	IngestEvent(ctx context.Context, event domain.WebhookEvent) error
}

type IngestionHandler struct {
	service IngestionServiceInterface
}

type ingestEventRequest struct {
	ID        string `json:"id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	TenantID  string `json:"tenant_id" binding:"required"`
	Payload   any    `json:"payload" binding:"required"`
}

func NewIngestionHandler(service IngestionServiceInterface) *IngestionHandler {
	return &IngestionHandler{service: service}
}

func (h *IngestionHandler) IngestEvent(c *gin.Context) {
	var req ingestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := domain.WebhookEvent{
		ID:        req.ID,
		EventType: req.EventType,
		TenantID:  req.TenantID,
		Payload:   req.Payload,
	}

	if err := h.service.IngestEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "event accepted"})
}
