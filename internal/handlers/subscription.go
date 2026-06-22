package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/subscription"
)

// SubscriptionService defines the interface needed by SubscriptionHandler.
type SubscriptionService interface {
	ActivateSubscription(ctx context.Context, tenantID, eventType, targetURL string) (string, error)
	DeactivateSubscription(ctx context.Context, tenantID, subscriptionID string) error
}

type SubscriptionHandler struct {
	Service SubscriptionService
}

type activateSubscriptionRequest struct {
	TenantID  string `json:"tenant_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	TargetURL string `json:"target_url" binding:"required"`
}

type deactivateSubscriptionRequest struct {
	TenantID       string `json:"tenant_id" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		Service: service,
	}
}

func (h *SubscriptionHandler) ActivateSubscription(c *gin.Context) {
	var req activateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secretKey, err := h.Service.ActivateSubscription(c.Request.Context(), req.TenantID, req.EventType, req.TargetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "subscription activated", "secretKey": secretKey})
}

func (h *SubscriptionHandler) DeactivateSubscription(c *gin.Context) {
	var req deactivateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.Service.DeactivateSubscription(c.Request.Context(), req.TenantID, req.SubscriptionID)
	if err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "subscription deactivated"})
}
