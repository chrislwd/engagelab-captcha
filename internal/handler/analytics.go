package handler

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

// CountryBreakdown handles GET /v1/analytics/countries
func (h *AnalyticsHandler) CountryBreakdown(c *gin.Context) {
	// In production, aggregated from challenge_sessions
	c.JSON(http.StatusOK, gin.H{
		"countries": []gin.H{
			{"code": "ID", "requests": 45200, "challenge_rate": 12.5, "pass_rate": 94.2, "bot_rate": 3.1},
			{"code": "TH", "requests": 32100, "challenge_rate": 15.3, "pass_rate": 91.8, "bot_rate": 4.5},
			{"code": "US", "requests": 28500, "challenge_rate": 8.2, "pass_rate": 96.5, "bot_rate": 1.8},
			{"code": "BR", "requests": 18900, "challenge_rate": 18.7, "pass_rate": 88.4, "bot_rate": 6.2},
			{"code": "IN", "requests": 15600, "challenge_rate": 22.1, "pass_rate": 85.3, "bot_rate": 8.9},
			{"code": "JP", "requests": 12300, "challenge_rate": 6.5, "pass_rate": 97.1, "bot_rate": 1.2},
		},
	})
}

// DeviceBreakdown handles GET /v1/analytics/devices
func (h *AnalyticsHandler) DeviceBreakdown(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"browsers": []gin.H{
			{"name": "Chrome", "pct": 62.3, "challenge_rate": 10.1},
			{"name": "Safari", "pct": 18.5, "challenge_rate": 7.8},
			{"name": "Firefox", "pct": 8.2, "challenge_rate": 9.5},
			{"name": "Edge", "pct": 5.1, "challenge_rate": 8.9},
			{"name": "Other", "pct": 5.9, "challenge_rate": 25.3},
		},
		"platforms": []gin.H{
			{"name": "Mobile", "pct": 58.2, "challenge_rate": 11.3},
			{"name": "Desktop", "pct": 38.5, "challenge_rate": 9.8},
			{"name": "Tablet", "pct": 3.3, "challenge_rate": 8.1},
		},
		"os": []gin.H{
			{"name": "Android", "pct": 42.1},
			{"name": "iOS", "pct": 22.8},
			{"name": "Windows", "pct": 24.3},
			{"name": "macOS", "pct": 8.5},
			{"name": "Linux", "pct": 2.3},
		},
	})
}

// ChallengeBreakdown handles GET /v1/analytics/challenges
func (h *AnalyticsHandler) ChallengeBreakdown(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"challenge_types": []gin.H{
			{"type": "invisible", "count": 85200, "pass_rate": 99.2, "avg_duration_ms": 0},
			{"type": "slider", "count": 12400, "pass_rate": 89.5, "avg_duration_ms": 3200},
			{"type": "click", "count": 5800, "pass_rate": 82.1, "avg_duration_ms": 5400},
			{"type": "puzzle", "count": 2100, "pass_rate": 78.3, "avg_duration_ms": 7800},
		},
		"risk_distribution": gin.H{
			"low":      68.5,
			"medium":   18.2,
			"high":     10.1,
			"critical": 3.2,
		},
		"hourly_volume": generateHourlyVolume(),
	})
}

// RiskTrends handles GET /v1/analytics/risk-trends
func (h *AnalyticsHandler) RiskTrends(c *gin.Context) {
	// 7-day trend data
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	var trends []gin.H
	for _, day := range days {
		trends = append(trends, gin.H{
			"day":            day,
			"total_requests": 15000 + rand.Intn(5000),
			"bot_blocked":    200 + rand.Intn(100),
			"challenges":     1500 + rand.Intn(500),
			"false_positive": 10 + rand.Intn(10),
			"avg_risk_score": 12.0 + rand.Float64()*8,
		})
	}
	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

func generateHourlyVolume() []gin.H {
	var hours []gin.H
	for h := 0; h < 24; h++ {
		base := 3000
		// Peak hours
		if h >= 9 && h <= 21 {
			base = 6000
		}
		hours = append(hours, gin.H{
			"hour":    h,
			"volume":  base + rand.Intn(2000),
			"blocked": rand.Intn(100),
		})
	}
	return hours
}
