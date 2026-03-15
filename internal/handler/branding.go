package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/service/challenge"
)

type BrandingHandler struct {
	manager *challenge.BrandManager
}

func NewBrandingHandler(manager *challenge.BrandManager) *BrandingHandler {
	return &BrandingHandler{manager: manager}
}

// Get handles GET /v1/apps/:id/branding
func (h *BrandingHandler) Get(c *gin.Context) {
	appID := c.Param("id")
	cfg := h.manager.Get(appID)
	c.JSON(http.StatusOK, cfg)
}

// Set handles PUT /v1/apps/:id/branding
func (h *BrandingHandler) Set(c *gin.Context) {
	appID := c.Param("id")
	var cfg challenge.BrandConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.AppID = appID
	h.manager.Set(&cfg)
	c.JSON(http.StatusOK, gin.H{
		"brand": cfg,
		"css":   cfg.GenerateCSS(),
	})
}

// Delete handles DELETE /v1/apps/:id/branding
func (h *BrandingHandler) Delete(c *gin.Context) {
	appID := c.Param("id")
	h.manager.Delete(appID)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// GetCSS handles GET /v1/apps/:id/branding/css (public, for SDK to fetch)
func (h *BrandingHandler) GetCSS(c *gin.Context) {
	appID := c.Param("id")
	cfg := h.manager.Get(appID)
	c.Data(http.StatusOK, "text/css; charset=utf-8", []byte(cfg.GenerateCSS()))
}
