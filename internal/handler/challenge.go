package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
	challengeEngine "github.com/engagelab/captcha/internal/service/challenge"
)

// ChallengeHandler handles challenge rendering and verification endpoints.
type ChallengeHandler struct {
	store     *repository.MemoryStore
	challenge *challengeEngine.Engine
}

// NewChallengeHandler creates a new ChallengeHandler.
func NewChallengeHandler(store *repository.MemoryStore, challenge *challengeEngine.Engine) *ChallengeHandler {
	return &ChallengeHandler{
		store:     store,
		challenge: challenge,
	}
}

// Render returns the challenge configuration for a given challenge ID.
// The client uses this data to display the appropriate CAPTCHA widget.
func (h *ChallengeHandler) Render(c *gin.Context) {
	var req model.RenderChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Look up the challenge session to determine the type.
	session, err := h.store.GetChallenge(req.ChallengeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	if session.Status != model.ChallengeStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "challenge already resolved", "status": session.Status})
		return
	}

	// Generate the challenge content.
	config := h.challenge.GenerateChallenge(session.ChallengeType)
	// Override the challenge ID so the client maps it back to the session.
	config.ChallengeID = session.ID

	c.JSON(http.StatusOK, config)
}

// Verify validates the user's answer to a challenge.
func (h *ChallengeHandler) Verify(c *gin.Context) {
	var req model.SubmitChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Look up the session.
	session, err := h.store.GetChallenge(req.ChallengeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	if session.Status != model.ChallengeStatusPending {
		c.JSON(http.StatusConflict, model.SubmitChallengeResponse{
			Success: false,
			Message: "challenge already resolved",
		})
		return
	}

	// Validate the answer.
	valid, token := h.challenge.ValidateChallenge(req.ChallengeID, req.Answer)

	if valid {
		_ = h.store.UpdateChallengeStatus(req.ChallengeID, model.ChallengeStatusPassed)
		c.JSON(http.StatusOK, model.SubmitChallengeResponse{
			Success: true,
			Token:   token,
			Message: "challenge passed",
		})
		return
	}

	_ = h.store.UpdateChallengeStatus(req.ChallengeID, model.ChallengeStatusFailed)
	c.JSON(http.StatusOK, model.SubmitChallengeResponse{
		Success: false,
		Message: "incorrect answer",
	})
}
