package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/service/sms"
)

// SMSCheckRequest is the payload for POST /v1/sms/check.
type SMSCheckRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	IP          string `json:"ip"`
	Fingerprint string `json:"fingerprint"`
	AppID       string `json:"app_id" binding:"required"`
}

// SMSCheckResponse is returned from POST /v1/sms/check.
type SMSCheckResponse struct {
	Allowed       bool   `json:"allowed"`
	Reason        string `json:"reason"`
	WaitSeconds   int    `json:"wait_seconds"`
	DailyCostCents int   `json:"daily_cost_cents"`
}

// SMSHandler handles SMS abuse check endpoints.
type SMSHandler struct {
	protector *sms.SMSProtector
}

// NewSMSHandler creates a new SMSHandler.
func NewSMSHandler(protector *sms.SMSProtector) *SMSHandler {
	return &SMSHandler{protector: protector}
}

// Check handles POST /v1/sms/check.
func (h *SMSHandler) Check(c *gin.Context) {
	var req SMSCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	ip := req.IP
	if ip == "" {
		ip = c.ClientIP()
	}

	result := h.protector.CheckSMSRequest(req.PhoneNumber, ip, req.Fingerprint)

	// If allowed, record the cost.
	if result.Allowed {
		h.protector.RecordSMSCost(req.AppID)
	}

	c.JSON(http.StatusOK, SMSCheckResponse{
		Allowed:        result.Allowed,
		Reason:         result.Reason,
		WaitSeconds:    result.WaitSeconds,
		DailyCostCents: h.protector.GetDailyCostCents(req.AppID),
	})
}
