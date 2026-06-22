package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/subscription"
)

type SubscriptionHandler struct {
	Service *subscription.SubscriptionService
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

func NewSubscriptionHandler(service *subscription.SubscriptionService) *SubscriptionHandler {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "subscription deactivated"})
}
