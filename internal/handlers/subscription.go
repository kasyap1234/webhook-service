package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kasyap1234/webhook-service/internal/subscription"
)

type SubscriptionHandler struct {
	Service *subscription.SubscriptionService
}

func NewSubscriptionHandler(service *subscription.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		Service: service,
	}
}

func (h *SubscriptionHandler) ActivateSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	payload := c.PostForm("payload")
	tenantID := c.PostForm("tenantID")
	eventType := c.PostForm("eventType")
	targetURL := c.PostForm("targetURL")
	secretKey, err := h.Service.ActivateSubscription(ctx, payload, tenantID, eventType, targetURL)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "subscription activated", "secretKey": secretKey})
}

func (h *SubscriptionHandler) DeactivateSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	subscriptionID := c.PostForm("subscriptionID")
	tenantID := c.PostForm("tenantID")
	err := h.Service.DeactivateSubscription(ctx, tenantID, subscriptionID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "subscription deactivated"})
}
