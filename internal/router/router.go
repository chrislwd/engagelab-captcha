package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/config"
	"github.com/engagelab/captcha/internal/handler"
	"github.com/engagelab/captcha/internal/metrics"
	"github.com/engagelab/captcha/internal/middleware"
	"github.com/engagelab/captcha/internal/repository"
	challengeEngine "github.com/engagelab/captcha/internal/service/challenge"
	policyEngine "github.com/engagelab/captcha/internal/service/policy"
	riskEngine "github.com/engagelab/captcha/internal/service/risk"
	verifyService "github.com/engagelab/captcha/internal/service/verify"
	"github.com/engagelab/captcha/internal/service/webhook"
)

func New(cfg *config.Config, store *repository.MemoryStore) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(metrics.RequestMetrics(metrics.Global))

	// Services
	risk := riskEngine.NewEngine()
	policy := policyEngine.NewEngine()
	challenge := challengeEngine.NewEngine(cfg.JWTSecret)
	verify := verifyService.NewService(store, cfg.JWTSecret)
	webhookStore := webhook.NewMemoryStore()
	_ = webhook.NewService(webhookStore)
	i18n := challengeEngine.NewI18n()

	// Handlers
	precheckH := handler.NewPrecheckHandler(store, risk, policy, challenge)
	challengeH := handler.NewChallengeHandler(store, challenge)
	siteVerifyH := handler.NewSiteVerifyHandler(verify)
	feedbackH := handler.NewFeedbackHandler(store)
	appH := handler.NewAppHandler(store)
	sceneH := handler.NewSceneHandler(store)
	statsH := handler.NewStatsHandler(store)
	webhookH := handler.NewWebhookHandler(webhookStore)
	analyticsH := handler.NewAnalyticsHandler()
	authH := handler.NewAuthHandler(store)
	threatsH := handler.NewThreatsHandler()

	// --- Health ---
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy", "service": "engagelab-captcha", "version": "1.2.0",
		})
	})

	// --- Metrics (Prometheus) ---
	r.GET("/metrics", metrics.Global.Handler())

	// --- SDK endpoints (site_key auth) ---
	sdk := r.Group("/v1")
	sdk.Use(middleware.SiteKeyAuth(store))
	{
		sdk.POST("/risk/precheck", precheckH.Handle)
		sdk.POST("/challenge/render", challengeH.Render)
		sdk.POST("/challenge/verify", challengeH.Verify)
	}

	// --- Public endpoints (no auth) ---
	r.POST("/v1/siteverify", siteVerifyH.Handle)
	r.GET("/v1/i18n/:lang", func(c *gin.Context) {
		lang := c.Param("lang")
		c.JSON(http.StatusOK, gin.H{
			"lang": lang, "translations": i18n.GetAll(lang), "supported": i18n.SupportedLanguages(),
		})
	})

	// --- Auth endpoints (no auth required) ---
	auth := r.Group("/v1/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// --- Console endpoints (API key auth) ---
	console := r.Group("/v1")
	console.Use(middleware.APIKeyAuth(store))
	{
		// Apps
		console.POST("/apps", appH.Create)
		console.GET("/apps", appH.List)
		console.GET("/apps/:id", appH.Get)
		console.DELETE("/apps/:id", appH.Delete)

		// Scenes
		console.POST("/scenes", sceneH.Create)
		console.GET("/scenes", sceneH.List)

		// Policies (read from store)
		console.GET("/policies", func(c *gin.Context) {
			policies := store.ListPolicies()
			c.JSON(http.StatusOK, gin.H{"policies": policies})
		})

		// Feedback
		console.POST("/events/feedback", feedbackH.Handle)

		// Stats
		console.GET("/stats/dashboard", statsH.Dashboard)

		// Webhooks
		console.POST("/webhooks", webhookH.Create)
		console.GET("/webhooks", webhookH.List)
		console.DELETE("/webhooks/:id", webhookH.Delete)

		// Analytics
		console.GET("/analytics/countries", analyticsH.CountryBreakdown)
		console.GET("/analytics/devices", analyticsH.DeviceBreakdown)
		console.GET("/analytics/challenges", analyticsH.ChallengeBreakdown)
		console.GET("/analytics/risk-trends", analyticsH.RiskTrends)

		// Threats / Attack Monitoring
		console.GET("/threats", threatsH.List)
		console.GET("/threats/dashboard", threatsH.Dashboard)
		console.POST("/threats/:id/mitigate", threatsH.Mitigate)

		// Account / API Keys
		console.POST("/account/api-keys", authH.GenerateAPIKey)
		console.GET("/account/api-keys", authH.ListAPIKeys)
	}

	return r
}
