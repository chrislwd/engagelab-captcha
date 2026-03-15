package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
	challengeEngine "github.com/engagelab/captcha/internal/service/challenge"
	policyEngine "github.com/engagelab/captcha/internal/service/policy"
	riskEngine "github.com/engagelab/captcha/internal/service/risk"
)

// PrecheckHandler handles the /v1/risk/precheck endpoint.
type PrecheckHandler struct {
	store     *repository.MemoryStore
	risk      *riskEngine.Engine
	policy    *policyEngine.Engine
	challenge *challengeEngine.Engine
}

// NewPrecheckHandler creates a new PrecheckHandler.
func NewPrecheckHandler(
	store *repository.MemoryStore,
	risk *riskEngine.Engine,
	policy *policyEngine.Engine,
	challenge *challengeEngine.Engine,
) *PrecheckHandler {
	return &PrecheckHandler{
		store:     store,
		risk:      risk,
		policy:    policy,
		challenge: challenge,
	}
}

// Handle processes a precheck request: runs risk analysis, evaluates policy,
// creates a challenge session, and returns the recommended action.
func (h *PrecheckHandler) Handle(c *gin.Context) {
	var req model.PrecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Resolve client IP: prefer the value in the request body, fall back to connection IP.
	clientIP := req.IP
	if clientIP == "" {
		clientIP = c.ClientIP()
	}

	// Validate app exists.
	app, err := h.store.GetApp(req.AppID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app_id"})
		return
	}

	// Look up the scene and its policy.
	var pol *model.Policy
	if req.SceneID != "" {
		scene, err := h.store.GetScene(req.SceneID)
		if err == nil && scene.PolicyID != "" {
			pol, _ = h.store.GetPolicy(scene.PolicyID)
		}
	}

	// Run risk assessment.
	riskResult := h.risk.Evaluate(clientIP, req.UserAgent, req.Fingerprint, req.BehaviorData)

	// Evaluate policy to get the final action.
	action := h.policy.Evaluate(riskResult.Score, pol, clientIP)

	// Determine challenge type.
	challengeType := h.policy.SelectChallengeType(action, riskResult.Score)

	// Create challenge session.
	sessionID := uuid.NewString()
	challengeSession := &model.ChallengeSession{
		ID:            uuid.NewString(),
		AppID:         app.ID,
		SceneID:       req.SceneID,
		SessionID:     sessionID,
		IP:            clientIP,
		UAHash:        hashUA(req.UserAgent),
		FingerprintID: req.Fingerprint,
		ChallengeType: challengeType,
		RiskScore:     riskResult.Score,
		RiskLabel:     riskResult.Label,
		Status:        model.ChallengeStatusPending,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}
	_ = h.store.SaveChallenge(challengeSession)

	resp := model.PrecheckResponse{
		Action:    action,
		RiskScore: riskResult.Score,
	}

	// For pass/invisible actions, generate a token immediately so the client does not
	// need to render a visual challenge.
	if action == model.RiskActionPass || action == model.RiskActionInvisible {
		cfg := h.challenge.GenerateChallenge(model.ChallengeTypeInvisible)
		resp.Token = cfg.Token
		resp.ChallengeID = challengeSession.ID
		resp.ChallengeType = model.ChallengeTypeInvisible

		// Mark session as passed immediately.
		_ = h.store.UpdateChallengeStatus(challengeSession.ID, model.ChallengeStatusPassed)
	} else if action == model.RiskActionChallenge {
		resp.ChallengeID = challengeSession.ID
		resp.ChallengeType = challengeType
	} else {
		// Deny: still provide a challenge ID so the client can show an error.
		resp.ChallengeID = challengeSession.ID
		resp.ChallengeType = challengeType
	}

	c.JSON(http.StatusOK, resp)
}

func hashUA(ua string) string {
	if ua == "" {
		return ""
	}
	h := sha256.Sum256([]byte(ua))
	return fmt.Sprintf("%x", h[:8])
}
