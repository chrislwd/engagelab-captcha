package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

// SceneHandler handles CRUD operations for scenes.
type SceneHandler struct {
	store *repository.MemoryStore
}

// NewSceneHandler creates a new SceneHandler.
func NewSceneHandler(store *repository.MemoryStore) *SceneHandler {
	return &SceneHandler{store: store}
}

// Create handles POST /v1/apps/:app_id/scenes.
func (h *SceneHandler) Create(c *gin.Context) {
	appID := c.Param("app_id")

	// Verify the app exists.
	_, err := h.store.GetApp(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	var req model.CreateSceneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate scene type.
	validTypes := map[model.SceneType]bool{
		model.SceneTypeRegister: true,
		model.SceneTypeLogin:    true,
		model.SceneTypeActivity: true,
		model.SceneTypeComment:  true,
		model.SceneTypeAPI:      true,
	}
	if !validTypes[req.SceneType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene_type, must be one of: register, login, activity, comment, api"})
		return
	}

	now := time.Now()
	scene := &model.Scene{
		ID:        uuid.NewString(),
		AppID:     appID,
		SceneType: req.SceneType,
		PolicyID:  req.PolicyID,
		Status:    model.SceneStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.store.CreateScene(scene); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create scene"})
		return
	}

	c.JSON(http.StatusCreated, model.CreateSceneResponse{
		ID:        scene.ID,
		AppID:     scene.AppID,
		SceneType: scene.SceneType,
		PolicyID:  scene.PolicyID,
		Status:    scene.Status,
		CreatedAt: scene.CreatedAt,
	})
}

// List handles GET /v1/apps/:app_id/scenes.
func (h *SceneHandler) List(c *gin.Context) {
	appID := c.Param("app_id")

	// Verify the app exists.
	_, err := h.store.GetApp(appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	scenes := h.store.ListScenesByApp(appID)
	if scenes == nil {
		scenes = []*model.Scene{}
	}

	c.JSON(http.StatusOK, gin.H{
		"scenes": scenes,
		"total":  len(scenes),
	})
}
