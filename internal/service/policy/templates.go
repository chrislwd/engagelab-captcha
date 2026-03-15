package policy

import (
	"time"

	"github.com/engagelab/captcha/internal/model"
)

// TemplateInfo describes an available industry policy template.
type TemplateInfo struct {
	Industry    string       `json:"industry"`
	Description string       `json:"description"`
	Policy      model.Policy `json:"policy"`
}

// templates holds pre-configured policies for common industries.
var templates = map[string]TemplateInfo{
	"ecommerce": {
		Industry:    "ecommerce",
		Description: "E-commerce: balanced protection for checkout, cart, and product pages",
		Policy: model.Policy{
			ID:            "template-ecommerce",
			SceneType:     model.SceneTypeActivity,
			ThresholdLow:  25,
			ThresholdHigh: 60,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  60,
			RateLimitRPH:  600,
			Enabled:       true,
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
		},
	},
	"fintech": {
		Industry:    "fintech",
		Description: "Fintech: strict protection for login, transactions, and account management",
		Policy: model.Policy{
			ID:            "template-fintech",
			SceneType:     model.SceneTypeLogin,
			ThresholdLow:  15,
			ThresholdHigh: 40,
			ActionLow:     model.RiskActionInvisible,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  30,
			RateLimitRPH:  200,
			Enabled:       true,
		},
	},
	"gaming": {
		Industry:    "gaming",
		Description: "Gaming: moderate protection to avoid friction for legitimate gamers",
		Policy: model.Policy{
			ID:            "template-gaming",
			SceneType:     model.SceneTypeLogin,
			ThresholdLow:  30,
			ThresholdHigh: 65,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  100,
			RateLimitRPH:  1000,
			Enabled:       true,
		},
	},
	"social": {
		Industry:    "social",
		Description: "Social media: focused on anti-spam for registration and posting",
		Policy: model.Policy{
			ID:            "template-social",
			SceneType:     model.SceneTypeRegister,
			ThresholdLow:  20,
			ThresholdHigh: 55,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  45,
			RateLimitRPH:  400,
			Enabled:       true,
		},
	},
	"education": {
		Industry:    "education",
		Description: "Education: lenient thresholds to minimize student friction",
		Policy: model.Policy{
			ID:            "template-education",
			SceneType:     model.SceneTypeLogin,
			ThresholdLow:  35,
			ThresholdHigh: 70,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionInvisible,
			ActionHigh:    model.RiskActionChallenge,
			RateLimitRPM:  80,
			RateLimitRPH:  800,
			Enabled:       true,
		},
	},
	"ticketing": {
		Industry:    "ticketing",
		Description: "Ticketing: aggressive bot prevention for high-demand ticket sales",
		Policy: model.Policy{
			ID:            "template-ticketing",
			SceneType:     model.SceneTypeActivity,
			ThresholdLow:  10,
			ThresholdHigh: 35,
			ActionLow:     model.RiskActionInvisible,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  20,
			RateLimitRPH:  100,
			Enabled:       true,
		},
	},
}

// GetTemplate returns a pre-configured policy for the given industry.
// Returns nil if the industry is not recognized.
func GetTemplate(industry string) *model.Policy {
	tmpl, ok := templates[industry]
	if !ok {
		return nil
	}
	// Return a copy so callers cannot modify the template.
	p := tmpl.Policy
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return &p
}

// ListTemplates returns all available industry templates.
func ListTemplates() []TemplateInfo {
	result := make([]TemplateInfo, 0, len(templates))
	for _, tmpl := range templates {
		result = append(result, tmpl)
	}
	return result
}
