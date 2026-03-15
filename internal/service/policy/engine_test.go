package policy

import (
	"testing"

	"github.com/engagelab/captcha/internal/model"
)

func newTestPolicy() *model.Policy {
	return &model.Policy{
		ID:            "test-policy",
		ThresholdLow:  20,
		ThresholdHigh: 60,
		ActionLow:     model.RiskActionPass,
		ActionMid:     model.RiskActionChallenge,
		ActionHigh:    model.RiskActionDeny,
		Enabled:       true,
	}
}

func TestEvaluate_LowScore_Pass(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()

	action := e.Evaluate(10, pol, "1.2.3.4")
	if action != model.RiskActionPass {
		t.Errorf("expected pass for low score, got %s", action)
	}
}

func TestEvaluate_MidScore_Challenge(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()

	action := e.Evaluate(40, pol, "1.2.3.4")
	if action != model.RiskActionChallenge {
		t.Errorf("expected challenge for mid score, got %s", action)
	}
}

func TestEvaluate_HighScore_Deny(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()

	action := e.Evaluate(80, pol, "1.2.3.4")
	if action != model.RiskActionDeny {
		t.Errorf("expected deny for high score, got %s", action)
	}
}

func TestEvaluate_BoundaryLow(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()

	// Exactly at low threshold should be pass
	action := e.Evaluate(20, pol, "1.2.3.4")
	if action != model.RiskActionPass {
		t.Errorf("expected pass at exact low threshold, got %s", action)
	}
}

func TestEvaluate_BoundaryHigh(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()

	// Exactly at high threshold should be challenge (mid range)
	action := e.Evaluate(60, pol, "1.2.3.4")
	if action != model.RiskActionChallenge {
		t.Errorf("expected challenge at exact high threshold, got %s", action)
	}
}

func TestEvaluate_IPWhitelist_OverridesToPass(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()
	pol.IPWhitelist = []string{"10.0.0.1", "192.168.0.0/16"}

	// Even with high risk score, whitelisted IP should pass
	action := e.Evaluate(90, pol, "10.0.0.1")
	if action != model.RiskActionPass {
		t.Errorf("expected pass for whitelisted IP, got %s", action)
	}

	// CIDR whitelist
	action = e.Evaluate(90, pol, "192.168.1.100")
	if action != model.RiskActionPass {
		t.Errorf("expected pass for IP in whitelisted CIDR, got %s", action)
	}
}

func TestEvaluate_IPBlacklist_OverridesToDeny(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()
	pol.IPBlacklist = []string{"10.0.0.99", "172.16.0.0/12"}

	// Even with low risk score, blacklisted IP should be denied
	action := e.Evaluate(5, pol, "10.0.0.99")
	if action != model.RiskActionDeny {
		t.Errorf("expected deny for blacklisted IP, got %s", action)
	}

	// CIDR blacklist
	action = e.Evaluate(5, pol, "172.16.5.10")
	if action != model.RiskActionDeny {
		t.Errorf("expected deny for IP in blacklisted CIDR, got %s", action)
	}
}

func TestEvaluate_WhitelistTakesPrecedenceOverBlacklist(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()
	pol.IPWhitelist = []string{"10.0.0.1"}
	pol.IPBlacklist = []string{"10.0.0.1"}

	// Whitelist is checked first
	action := e.Evaluate(50, pol, "10.0.0.1")
	if action != model.RiskActionPass {
		t.Errorf("expected whitelist to take precedence, got %s", action)
	}
}

func TestEvaluate_DefaultPolicyWhenNil(t *testing.T) {
	e := NewEngine()

	// No policy: fall back to score-based defaults
	tests := []struct {
		score    float64
		expected model.RiskAction
	}{
		{10, model.RiskActionPass},
		{30, model.RiskActionInvisible},
		{60, model.RiskActionChallenge},
		{90, model.RiskActionDeny},
	}

	for _, tt := range tests {
		action := e.Evaluate(tt.score, nil, "1.2.3.4")
		if action != tt.expected {
			t.Errorf("defaultAction(%f) = %q, want %q", tt.score, action, tt.expected)
		}
	}
}

func TestEvaluate_DisabledPolicyUsesDefaults(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()
	pol.Enabled = false

	action := e.Evaluate(10, pol, "1.2.3.4")
	if action != model.RiskActionPass {
		t.Errorf("expected default pass for disabled policy with low score, got %s", action)
	}
}

func TestSelectChallengeType(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		action   model.RiskAction
		score    float64
		expected model.ChallengeType
	}{
		{model.RiskActionPass, 5, model.ChallengeTypeInvisible},
		{model.RiskActionInvisible, 25, model.ChallengeTypeInvisible},
		{model.RiskActionChallenge, 40, model.ChallengeTypeSlider},
		{model.RiskActionChallenge, 65, model.ChallengeTypeClick},
		{model.RiskActionDeny, 90, model.ChallengeTypePuzzle},
	}

	for _, tt := range tests {
		result := e.SelectChallengeType(tt.action, tt.score)
		if result != tt.expected {
			t.Errorf("SelectChallengeType(%s, %f) = %s, want %s", tt.action, tt.score, result, tt.expected)
		}
	}
}

func TestEvaluate_EmptyClientIP(t *testing.T) {
	e := NewEngine()
	pol := newTestPolicy()
	pol.IPWhitelist = []string{"10.0.0.1"}

	// Empty IP should not match whitelist
	action := e.Evaluate(50, pol, "")
	if action != model.RiskActionChallenge {
		t.Errorf("expected challenge for empty IP with mid score, got %s", action)
	}
}

func TestEvaluate_CoalesceEmptyActions(t *testing.T) {
	e := NewEngine()
	pol := &model.Policy{
		ID:            "test-empty-actions",
		ThresholdLow:  20,
		ThresholdHigh: 60,
		ActionLow:     "", // empty - should use fallback
		ActionMid:     "", // empty - should use fallback
		ActionHigh:    "", // empty - should use fallback
		Enabled:       true,
	}

	if action := e.Evaluate(10, pol, "1.2.3.4"); action != model.RiskActionPass {
		t.Errorf("expected fallback pass, got %s", action)
	}
	if action := e.Evaluate(40, pol, "1.2.3.4"); action != model.RiskActionChallenge {
		t.Errorf("expected fallback challenge, got %s", action)
	}
	if action := e.Evaluate(80, pol, "1.2.3.4"); action != model.RiskActionDeny {
		t.Errorf("expected fallback deny, got %s", action)
	}
}
