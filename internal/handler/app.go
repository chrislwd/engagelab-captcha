package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

// AppHandler handles CRUD operations for apps.
type AppHandler struct {
	store *repository.MemoryStore
}

// NewAppHandler creates a new AppHandler.
func NewAppHandler(store *repository.MemoryStore) *AppHandler {
	return &AppHandler{store: store}
}

// Create handles POST /v1/apps.
func (h *AppHandler) Create(c *gin.Context) {
	var req model.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Get tenant ID from context (set by auth middleware).
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)
	if tid == "" {
		tid = "tenant-001" // fallback for development
	}

	now := time.Now()
	app := &model.App{
		ID:             uuid.NewString(),
		TenantID:       tid,
		Name:           req.Name,
		SiteKey:        "sk_" + uuid.NewString()[:20],
		SecretKey:      "sec_" + uuid.NewString()[:24],
		AllowedDomains: req.AllowedDomains,
		Status:         model.AppStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if app.AllowedDomains == nil {
		app.AllowedDomains = []string{}
	}

	if err := h.store.CreateApp(app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create app"})
		return
	}

	c.JSON(http.StatusCreated, model.CreateAppResponse{
		ID:        app.ID,
		Name:      app.Name,
		SiteKey:   app.SiteKey,
		SecretKey: app.SecretKey,
		Status:    app.Status,
		CreatedAt: app.CreatedAt,
	})
}

// List handles GET /v1/apps.
func (h *AppHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)
	if tid == "" {
		tid = "tenant-001"
	}

	apps := h.store.ListAppsByTenant(tid)
	if apps == nil {
		apps = []*model.App{}
	}

	c.JSON(http.StatusOK, gin.H{
		"apps":  apps,
		"total": len(apps),
	})
}

// Get handles GET /v1/apps/:id.
func (h *AppHandler) Get(c *gin.Context) {
	id := c.Param("id")
	app, err := h.store.GetApp(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

// Delete handles DELETE /v1/apps/:id.
func (h *AppHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteApp(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "app deleted"})
}
