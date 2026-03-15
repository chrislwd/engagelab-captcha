package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

// FeedbackHandler handles the /v1/events/feedback endpoint.
type FeedbackHandler struct {
	store *repository.MemoryStore
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(store *repository.MemoryStore) *FeedbackHandler {
	return &FeedbackHandler{store: store}
}

// Handle processes event feedback submissions (false positives, false negatives).
func (h *FeedbackHandler) Handle(c *gin.Context) {
	var req model.EventFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate that the challenge exists.
	_, err := h.store.GetChallenge(req.ChallengeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	// Validate feedback type.
	validTypes := map[string]bool{
		"false_positive": true,
		"false_negative": true,
		"abuse":          true,
		"other":          true,
	}
	if !validTypes[req.FeedbackType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feedback_type, must be one of: false_positive, false_negative, abuse, other"})
		return
	}

	fb := &model.EventFeedback{
		ID:           uuid.NewString(),
		ChallengeID:  req.ChallengeID,
		FeedbackType: req.FeedbackType,
		Comment:      req.Comment,
		CreatedAt:    time.Now(),
	}

	if err := h.store.SaveFeedback(fb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      fb.ID,
		"status":  "received",
		"message": "feedback recorded successfully",
	})
}
