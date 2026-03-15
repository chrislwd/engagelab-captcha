package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

// StatsHandler handles the /v1/stats/dashboard endpoint.
type StatsHandler struct {
	store *repository.MemoryStore
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(store *repository.MemoryStore) *StatsHandler {
	return &StatsHandler{store: store}
}

// Dashboard handles GET /v1/stats/dashboard.
// It computes aggregate statistics from all challenge sessions.
func (h *StatsHandler) Dashboard(c *gin.Context) {
	challenges := h.store.ListChallenges()

	total := int64(len(challenges))
	var passed, failed, denied, expired int64

	riskDist := map[string]int64{
		"low":      0,
		"medium":   0,
		"high":     0,
		"critical": 0,
	}

	// Simple country simulation based on IP prefix for demo purposes.
	countryMap := map[string]int64{}

	for _, ch := range challenges {
		switch ch.Status {
		case model.ChallengeStatusPassed:
			passed++
		case model.ChallengeStatusFailed:
			failed++
		case model.ChallengeStatusExpired:
			expired++
		}

		// Classify risk score.
		switch {
		case ch.RiskScore <= 15:
			riskDist["low"]++
		case ch.RiskScore <= 40:
			riskDist["medium"]++
		case ch.RiskScore <= 70:
			riskDist["high"]++
		default:
			riskDist["critical"]++
		}

		// Derive country from IP for demo. In production this would use a GeoIP database.
		country := deriveCountry(ch.IP)
		countryMap[country]++
	}

	// Count denied as failed + expired for rate purposes.
	denied = failed + expired

	// Compute rates safely.
	var challengeRate, passRate, denyRate float64
	if total > 0 {
		challengeRate = float64(total)
		passRate = float64(passed) / float64(total) * 100
		denyRate = float64(denied) / float64(total) * 100
	}

	// Build top countries list.
	var topCountries []model.CountryStat
	for country, count := range countryMap {
		topCountries = append(topCountries, model.CountryStat{
			Country: country,
			Count:   count,
		})
	}

	stats := model.DashboardStats{
		TotalChallenges:  total,
		ChallengeRate:    challengeRate,
		PassRate:         passRate,
		DenyRate:         denyRate,
		RiskDistribution: riskDist,
		TopCountries:     topCountries,
	}

	c.JSON(http.StatusOK, stats)
}

// deriveCountry maps IP prefixes to countries for demo purposes.
func deriveCountry(ip string) string {
	if len(ip) < 4 {
		return "Unknown"
	}
	switch {
	case len(ip) >= 4 && ip[:4] == "192.":
		return "US"
	case len(ip) >= 3 && ip[:3] == "10.":
		return "CN"
	case len(ip) >= 4 && ip[:4] == "172.":
		return "DE"
	case len(ip) >= 3 && ip[:3] == "35.":
		return "US"
	case len(ip) >= 3 && ip[:3] == "52.":
		return "JP"
	default:
		return "Other"
	}
}
