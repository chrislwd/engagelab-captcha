package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/model"
	verifyService "github.com/engagelab/captcha/internal/service/verify"
)

// SiteVerifyHandler handles the /v1/siteverify endpoint.
type SiteVerifyHandler struct {
	verify *verifyService.Service
}

// NewSiteVerifyHandler creates a new SiteVerifyHandler.
func NewSiteVerifyHandler(verify *verifyService.Service) *SiteVerifyHandler {
	return &SiteVerifyHandler{verify: verify}
}

// Handle validates a CAPTCHA token submitted by the customer's backend.
func (h *SiteVerifyHandler) Handle(c *gin.Context) {
	var req model.SiteVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-request"},
		})
		return
	}

	resp := h.verify.SiteVerify(req)
	if resp.Success {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}
