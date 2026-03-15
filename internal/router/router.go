package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/config"
	"github.com/engagelab/captcha/internal/handler"
	"github.com/engagelab/captcha/internal/middleware"
	"github.com/engagelab/captcha/internal/repository"
	challengeEngine "github.com/engagelab/captcha/internal/service/challenge"
	policyEngine "github.com/engagelab/captcha/internal/service/policy"
	riskEngine "github.com/engagelab/captcha/internal/service/risk"
	verifyService "github.com/engagelab/captcha/internal/service/verify"
)

// New creates and configures the Gin router with all routes, handlers, and middleware.
func New(cfg *config.Config, store *repository.MemoryStore) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Initialize service engines.
	risk := riskEngine.NewEngine()
	policy := policyEngine.NewEngine()
	challenge := challengeEngine.NewEngine(cfg.JWTSecret)
	verify := verifyService.NewService(store, cfg.JWTSecret)

	// Initialize handlers.
	precheckH := handler.NewPrecheckHandler(store, risk, policy, challenge)
	challengeH := handler.NewChallengeHandler(store, challenge)
	siteVerifyH := handler.NewSiteVerifyHandler(verify)
	feedbackH := handler.NewFeedbackHandler(store)
	appH := handler.NewAppHandler(store)
	sceneH := handler.NewSceneHandler(store)
	statsH := handler.NewStatsHandler(store)

	// Health check endpoint (no auth).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "engagelab-captcha",
			"version": "1.0.0",
		})
	})

	// ----- SDK endpoints (site_key auth) -----
	sdk := r.Group("/v1")
	sdk.Use(middleware.SiteKeyAuth(store))
	{
		sdk.POST("/risk/precheck", precheckH.Handle)
		sdk.POST("/challenge/render", challengeH.Render)
		sdk.POST("/challenge/verify", challengeH.Verify)
	}

	// ----- Server-side verification (secret key in body, no middleware auth) -----
	r.POST("/v1/siteverify", siteVerifyH.Handle)

	// ----- Console/management endpoints (API key auth) -----
	console := r.Group("/v1")
	console.Use(middleware.APIKeyAuth(store))
	{
		// Apps CRUD.
		console.POST("/apps", appH.Create)
		console.GET("/apps", appH.List)
		console.GET("/apps/:id", appH.Get)
		console.DELETE("/apps/:id", appH.Delete)

		// Scenes CRUD.
		console.POST("/scenes", sceneH.Create)
		console.GET("/scenes", sceneH.List)

		// Feedback.
		console.POST("/events/feedback", feedbackH.Handle)

		// Dashboard stats.
		console.GET("/stats/dashboard", statsH.Dashboard)
	}

	return r
}
