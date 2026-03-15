package risk

import (
	"testing"

	"github.com/engagelab/captcha/internal/model"
)

func TestEvaluate_NormalBehavior_LowRisk(t *testing.T) {
	e := NewEngine()

	behavior := map[string]interface{}{
		"mouse_entropy":   3.5,
		"time_on_page_ms": 15000.0,
		"key_count":       42.0,
		"scroll_count":    8.0,
	}

	result := e.Evaluate("1.2.3.4", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "fp-normal-user-1234", behavior)

	if result.Score > 15 {
		t.Errorf("expected low risk score (<=15), got %f", result.Score)
	}
	if result.Action != model.RiskActionPass {
		t.Errorf("expected action pass, got %s", result.Action)
	}
}

func TestEvaluate_BotBehavior_HighRisk(t *testing.T) {
	e := NewEngine()

	behavior := map[string]interface{}{
		"mouse_entropy":   0.0,
		"time_on_page_ms": 100.0,
		"key_count":       0.0,
		"scroll_count":    0.0,
	}

	result := e.Evaluate("1.2.3.4", "", "", behavior)

	// Missing UA = 12, missing fingerprint = 5, low entropy = 15, instant = 15, no keys = 5, no scroll = 3 = 55+
	if result.Score < 40 {
		t.Errorf("expected high risk score (>=40), got %f", result.Score)
	}
}

func TestEvaluate_BotUserAgent(t *testing.T) {
	e := NewEngine()

	botUAs := []string{
		"python-requests/2.28.0",
		"curl/7.88.1",
		"Go-http-client/1.1",
		"Wget/1.21",
		"Puppeteer/19.0",
		"HeadlessChrome/112.0",
	}

	for _, ua := range botUAs {
		result := e.Evaluate("1.2.3.4", ua, "fp-test", nil)
		// bot_ua = 15, no_behavior_data = 10, at minimum
		if result.Score < 15 {
			t.Errorf("expected score >= 15 for bot UA %q, got %f", ua, result.Score)
		}
		if !containsSubstring(result.Label, "bot_ua") {
			t.Errorf("expected label to contain 'bot_ua' for UA %q, got %q", ua, result.Label)
		}
	}
}

func TestEvaluate_DatacenterIP(t *testing.T) {
	e := NewEngine()

	// 52.0.0.1 falls in the AWS 52.0.0.0/11 range
	result := e.Evaluate("52.0.0.1", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "fp-test", map[string]interface{}{
		"mouse_entropy":   3.0,
		"time_on_page_ms": 5000.0,
	})

	if !containsSubstring(result.Label, "datacenter_ip") {
		t.Errorf("expected label to contain 'datacenter_ip', got %q", result.Label)
	}
	// datacenter adds 15 points
	if result.Score < 15 {
		t.Errorf("expected score >= 15 for datacenter IP, got %f", result.Score)
	}
}

func TestEvaluate_PrivateIP(t *testing.T) {
	e := NewEngine()

	result := e.Evaluate("192.168.1.1", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "fp-test", map[string]interface{}{
		"mouse_entropy":   3.0,
		"time_on_page_ms": 5000.0,
	})

	// Private IP adds 2 points -- a minor signal
	if result.Score < 2 {
		t.Errorf("expected score >= 2 for private IP, got %f", result.Score)
	}
}

func TestEvaluate_RateLimiting(t *testing.T) {
	e := NewEngine()

	behavior := map[string]interface{}{
		"mouse_entropy":   3.0,
		"time_on_page_ms": 5000.0,
	}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// Send 20 requests from the same IP/fingerprint
	var lastResult Result
	for i := 0; i < 20; i++ {
		lastResult = e.Evaluate("10.0.0.1", ua, "same-fingerprint", behavior)
	}

	// After 20 requests in quick succession, rate score should be elevated
	if !containsSubstring(lastResult.Label, "rate") {
		t.Errorf("expected rate-related label after many requests, got %q", lastResult.Label)
	}
	if lastResult.Score < 8 {
		t.Errorf("expected elevated score from rate limiting, got %f", lastResult.Score)
	}
}

func TestEvaluate_ScoreClampedTo100(t *testing.T) {
	e := NewEngine()

	// Blacklisted IP (30) + datacenter (15) + missing UA (12) + missing fingerprint (5) +
	// no behavior (10) = 72, but with rate limiting from repeated calls...
	// Use a known bad IP that is also in datacenter range
	// 198.51.100.50 is both a bad IP and in the 198.51.100.0/24 range
	for i := 0; i < 40; i++ {
		e.Evaluate("198.51.100.50", "", "", nil)
	}
	result := e.Evaluate("198.51.100.50", "", "", nil)

	if result.Score > 100 {
		t.Errorf("score should be clamped to 100, got %f", result.Score)
	}
	if result.Score < 0 {
		t.Errorf("score should not be negative, got %f", result.Score)
	}
}

func TestEvaluate_ScoreNotNegative(t *testing.T) {
	e := NewEngine()

	result := e.Evaluate("1.2.3.4", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", "good-fp", map[string]interface{}{
		"mouse_entropy":   4.0,
		"time_on_page_ms": 30000.0,
		"key_count":       100.0,
		"scroll_count":    20.0,
	})

	if result.Score < 0 {
		t.Errorf("score should not be negative, got %f", result.Score)
	}
}

func TestClassifyRisk(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{0, "low"},
		{15, "low"},
		{16, "medium"},
		{40, "medium"},
		{41, "high"},
		{70, "high"},
		{71, "critical"},
		{100, "critical"},
	}

	for _, tt := range tests {
		got := classifyRisk(tt.score)
		if got != tt.expected {
			t.Errorf("classifyRisk(%f) = %q, want %q", tt.score, got, tt.expected)
		}
	}
}

func TestRecommendAction(t *testing.T) {
	tests := []struct {
		score    float64
		expected model.RiskAction
	}{
		{0, model.RiskActionPass},
		{15, model.RiskActionPass},
		{16, model.RiskActionInvisible},
		{30, model.RiskActionInvisible},
		{31, model.RiskActionChallenge},
		{70, model.RiskActionChallenge},
		{71, model.RiskActionDeny},
		{100, model.RiskActionDeny},
	}

	for _, tt := range tests {
		got := recommendAction(tt.score)
		if got != tt.expected {
			t.Errorf("recommendAction(%f) = %q, want %q", tt.score, got, tt.expected)
		}
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
