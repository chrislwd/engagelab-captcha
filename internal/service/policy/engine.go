package policy

import (
	"net"
	"strings"

	"github.com/engagelab/captcha/internal/model"
)

// Engine evaluates a risk score against a policy to determine the final action.
type Engine struct{}

// NewEngine creates a new policy evaluation engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Evaluate determines the appropriate action based on risk score and policy rules.
func (e *Engine) Evaluate(riskScore float64, policy *model.Policy, clientIP string) model.RiskAction {
	if policy == nil || !policy.Enabled {
		// No active policy: fall back to score-based defaults.
		return defaultAction(riskScore)
	}

	// 1. Check IP whitelist first (always pass).
	if len(policy.IPWhitelist) > 0 && matchesIPList(clientIP, policy.IPWhitelist) {
		return model.RiskActionPass
	}

	// 2. Check IP blacklist (always deny).
	if len(policy.IPBlacklist) > 0 && matchesIPList(clientIP, policy.IPBlacklist) {
		return model.RiskActionDeny
	}

	// 3. Threshold-based action selection.
	switch {
	case riskScore <= policy.ThresholdLow:
		return coalesce(policy.ActionLow, model.RiskActionPass)
	case riskScore <= policy.ThresholdHigh:
		return coalesce(policy.ActionMid, model.RiskActionChallenge)
	default:
		return coalesce(policy.ActionHigh, model.RiskActionDeny)
	}
}

// SelectChallengeType picks a challenge type based on the action and risk level.
func (e *Engine) SelectChallengeType(action model.RiskAction, riskScore float64) model.ChallengeType {
	switch action {
	case model.RiskActionPass:
		return model.ChallengeTypeInvisible
	case model.RiskActionInvisible:
		return model.ChallengeTypeInvisible
	case model.RiskActionChallenge:
		if riskScore > 60 {
			return model.ChallengeTypeClick
		}
		return model.ChallengeTypeSlider
	case model.RiskActionDeny:
		return model.ChallengeTypePuzzle
	default:
		return model.ChallengeTypeSlider
	}
}

// matchesIPList checks if an IP matches any entry in a list.
// Entries can be single IPs or CIDR notation.
func matchesIPList(clientIP string, list []string) bool {
	if clientIP == "" {
		return false
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, entry := range list {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Try CIDR match.
		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err == nil && cidr.Contains(ip) {
				return true
			}
			continue
		}

		// Exact IP match.
		if entry == clientIP {
			return true
		}
	}

	return false
}

func defaultAction(score float64) model.RiskAction {
	switch {
	case score <= 20:
		return model.RiskActionPass
	case score <= 50:
		return model.RiskActionInvisible
	case score <= 75:
		return model.RiskActionChallenge
	default:
		return model.RiskActionDeny
	}
}

func coalesce(action model.RiskAction, fallback model.RiskAction) model.RiskAction {
	if action == "" {
		return fallback
	}
	return action
}
