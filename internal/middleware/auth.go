package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/repository"
)

// APIKeyAuth returns middleware that authenticates requests using the X-API-Key header.
// Used for console/management API endpoints.
func APIKeyAuth(store *repository.MemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-API-Key header",
			})
			return
		}

		tenant, err := store.GetTenantByAPIKey(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		// Store tenant info in context for downstream handlers.
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_name", tenant.Name)
		c.Set("tenant_plan", string(tenant.Plan))
		c.Next()
	}
}

// SiteKeyAuth returns middleware that authenticates SDK requests using a site key.
// The site key can be provided via the X-Site-Key header or as an "app_id" field in the request body.
// This is a lighter auth mechanism for client-side SDK calls.
func SiteKeyAuth(store *repository.MemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteKey := c.GetHeader("X-Site-Key")

		// Also check Authorization Bearer for SDK usage.
		if siteKey == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				siteKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if siteKey != "" {
			app, err := store.GetAppBySiteKey(siteKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "invalid site key",
				})
				return
			}
			c.Set("app_id", app.ID)
			c.Set("site_key", app.SiteKey)
			c.Set("tenant_id", app.TenantID)
		}
		// If no site key is provided, allow the request through.
		// The handler will validate the app_id from the request body.
		c.Next()
	}
}
