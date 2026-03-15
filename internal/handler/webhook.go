package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/engagelab/captcha/internal/service/webhook"
)

type WebhookHandler struct {
	store *webhook.MemoryWebhookStore
}

func NewWebhookHandler(store *webhook.MemoryWebhookStore) *WebhookHandler {
	return &WebhookHandler{store: store}
}

// Create handles POST /v1/webhooks
func (h *WebhookHandler) Create(c *gin.Context) {
	var req webhook.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &webhook.Subscription{
		ID:     uuid.New().String(),
		AppID:  req.AppID,
		URL:    req.URL,
		Secret: req.Secret,
		Events: req.Events,
		Active: true,
	}

	h.store.Add(sub)
	c.JSON(http.StatusCreated, webhook.FormatResponse(sub))
}

// List handles GET /v1/webhooks
func (h *WebhookHandler) List(c *gin.Context) {
	subs := h.store.List()
	c.JSON(http.StatusOK, gin.H{"webhooks": webhook.FormatList(subs)})
}

// Delete handles DELETE /v1/webhooks/:id
func (h *WebhookHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if h.store.Delete(id) {
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
	}
}

// Unused import guard
var _ = time.Now
