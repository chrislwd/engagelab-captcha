package challenge

import (
	"strings"
	"testing"

	"github.com/engagelab/captcha/internal/model"
)

const testHMACSecret = "test-secret-key-for-hmac"

func TestGenerateSliderChallenge(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	if cfg.ChallengeID == "" {
		t.Error("expected non-empty challenge ID")
	}
	if cfg.ChallengeType != model.ChallengeTypeSlider {
		t.Errorf("expected challenge type slider, got %s", cfg.ChallengeType)
	}
	if cfg.SliderBgWidth == 0 || cfg.SliderBgHeight == 0 {
		t.Error("expected non-zero slider background dimensions")
	}
	if cfg.SliderWidth == 0 || cfg.SliderHeight == 0 {
		t.Error("expected non-zero slider piece dimensions")
	}
	if cfg.ExpiresAt.IsZero() {
		t.Error("expected non-zero expiration time")
	}
}

func TestGenerateClickChallenge(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeClick)

	if cfg.ChallengeID == "" {
		t.Error("expected non-empty challenge ID")
	}
	if cfg.ChallengeType != model.ChallengeTypeClick {
		t.Errorf("expected challenge type click, got %s", cfg.ChallengeType)
	}
	if len(cfg.ClickTargets) == 0 {
		t.Error("expected non-empty click targets")
	}
	for _, target := range cfg.ClickTargets {
		if target.ID == "" {
			t.Error("expected non-empty target ID")
		}
	}
	if cfg.ClickPrompt == "" {
		t.Error("expected non-empty click prompt")
	}
}

func TestGenerateInvisibleChallenge(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeInvisible)

	if cfg.ChallengeID == "" {
		t.Error("expected non-empty challenge ID")
	}
	if cfg.ChallengeType != model.ChallengeTypeInvisible {
		t.Errorf("expected challenge type invisible, got %s", cfg.ChallengeType)
	}
	if cfg.Token == "" {
		t.Error("expected non-empty token for invisible challenge")
	}
	// Token format: payload.signature
	parts := strings.SplitN(cfg.Token, ".", 2)
	if len(parts) != 2 {
		t.Errorf("expected token format 'payload.signature', got %q", cfg.Token)
	}
}

func TestValidateSlider_CorrectPosition(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	// Read the stored answer to know the target X
	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(sliderAnswer)
	e.mu.RUnlock()

	// Submit the exact target X
	valid, token := e.ValidateChallenge(cfg.ChallengeID, float64(expected.TargetX))
	if !valid {
		t.Error("expected valid result for correct slider position")
	}
	if token == "" {
		t.Error("expected non-empty token on successful validation")
	}
}

func TestValidateSlider_WithinTolerance(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(sliderAnswer)
	e.mu.RUnlock()

	// Submit position within tolerance
	valid, token := e.ValidateChallenge(cfg.ChallengeID, float64(expected.TargetX+expected.Tolerance))
	if !valid {
		t.Error("expected valid result for position within tolerance")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateSlider_WrongPosition(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(sliderAnswer)
	e.mu.RUnlock()

	// Submit position far from target
	valid, token := e.ValidateChallenge(cfg.ChallengeID, float64(expected.TargetX+50))
	if valid {
		t.Error("expected invalid result for wrong slider position")
	}
	if token != "" {
		t.Error("expected empty token on failed validation")
	}
}

func TestValidateSlider_MapAnswer(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(sliderAnswer)
	e.mu.RUnlock()

	// Submit as map with "x" key
	answer := map[string]interface{}{"x": float64(expected.TargetX)}
	valid, token := e.ValidateChallenge(cfg.ChallengeID, answer)
	if !valid {
		t.Error("expected valid result for correct slider position via map answer")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateClick_CorrectOrder(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeClick)

	// Read the stored answer
	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(clickAnswer)
	correctIDs := make([]string, len(expected.TargetIDs))
	copy(correctIDs, expected.TargetIDs)
	e.mu.RUnlock()

	// Submit correct IDs as []interface{}
	answer := make([]interface{}, len(correctIDs))
	for i, id := range correctIDs {
		answer[i] = id
	}

	valid, token := e.ValidateChallenge(cfg.ChallengeID, answer)
	if !valid {
		t.Error("expected valid result for correct click targets")
	}
	if token == "" {
		t.Error("expected non-empty token on successful validation")
	}
}

func TestValidateClick_WrongTargets(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeClick)

	// Submit wrong IDs
	answer := []interface{}{"wrong-id-1", "wrong-id-2", "wrong-id-3"}
	valid, token := e.ValidateChallenge(cfg.ChallengeID, answer)
	if valid {
		t.Error("expected invalid result for wrong click targets")
	}
	if token != "" {
		t.Error("expected empty token on failed validation")
	}
}

func TestValidateClick_WrongCount(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeClick)

	// Submit too few IDs
	answer := []interface{}{"some-id"}
	valid, _ := e.ValidateChallenge(cfg.ChallengeID, answer)
	if valid {
		t.Error("expected invalid result for wrong number of click targets")
	}
}

func TestTokenSignatureVerification(t *testing.T) {
	e := NewEngine(testHMACSecret)
	token := e.signToken("test-challenge-id")

	// Token format: challengeID|timestamp.signature
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		t.Fatalf("expected token format 'payload.signature', got %q", token)
	}

	payload := parts[0]
	if !strings.Contains(payload, "test-challenge-id") {
		t.Errorf("expected payload to contain challenge ID, got %q", payload)
	}
	if !strings.Contains(payload, "|") {
		t.Errorf("expected payload to contain pipe separator, got %q", payload)
	}

	// Verify different secrets produce different tokens
	e2 := NewEngine("different-secret")
	token2 := e2.signToken("test-challenge-id")

	sig1 := strings.SplitN(token, ".", 2)[1]
	sig2 := strings.SplitN(token2, ".", 2)[1]
	if sig1 == sig2 {
		t.Error("expected different secrets to produce different signatures")
	}
}

func TestReplayPrevention(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeSlider)

	e.mu.RLock()
	rec := e.answers[cfg.ChallengeID]
	expected := rec.Data.(sliderAnswer)
	e.mu.RUnlock()

	// First validation should succeed
	valid, _ := e.ValidateChallenge(cfg.ChallengeID, float64(expected.TargetX))
	if !valid {
		t.Error("expected first validation to succeed")
	}

	// Second validation of same challenge should fail (answer was deleted)
	valid2, token2 := e.ValidateChallenge(cfg.ChallengeID, float64(expected.TargetX))
	if valid2 {
		t.Error("expected second validation of same challenge to fail (replay prevention)")
	}
	if token2 != "" {
		t.Error("expected empty token on replay attempt")
	}
}

func TestValidateChallenge_UnknownID(t *testing.T) {
	e := NewEngine(testHMACSecret)

	valid, token := e.ValidateChallenge("nonexistent-id", float64(100))
	if valid {
		t.Error("expected invalid result for unknown challenge ID")
	}
	if token != "" {
		t.Error("expected empty token for unknown challenge ID")
	}
}

func TestValidateInvisible(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypeInvisible)

	valid, token := e.ValidateChallenge(cfg.ChallengeID, nil)
	if !valid {
		t.Error("expected invisible challenge to always validate")
	}
	if token == "" {
		t.Error("expected non-empty token for invisible challenge")
	}
}

func TestGeneratePuzzleChallenge(t *testing.T) {
	e := NewEngine(testHMACSecret)
	cfg := e.GenerateChallenge(model.ChallengeTypePuzzle)

	if cfg.ChallengeID == "" {
		t.Error("expected non-empty challenge ID")
	}
	if cfg.ChallengeType != model.ChallengeTypePuzzle {
		t.Errorf("expected challenge type puzzle, got %s", cfg.ChallengeType)
	}
	if cfg.PuzzleImageURL == "" {
		t.Error("expected non-empty puzzle image URL")
	}
}

func TestCleanExpired(t *testing.T) {
	e := NewEngine(testHMACSecret)

	// Generate some challenges
	cfg1 := e.GenerateChallenge(model.ChallengeTypeSlider)
	cfg2 := e.GenerateChallenge(model.ChallengeTypeSlider)

	// Manually expire one
	e.mu.Lock()
	if rec, ok := e.answers[cfg1.ChallengeID]; ok {
		rec.ExpiresAt = rec.ExpiresAt.Add(-10 * 60 * 1e9) // expired
		e.answers[cfg1.ChallengeID] = rec
	}
	e.mu.Unlock()

	e.CleanExpired()

	e.mu.RLock()
	_, exists1 := e.answers[cfg1.ChallengeID]
	_, exists2 := e.answers[cfg2.ChallengeID]
	e.mu.RUnlock()

	if exists1 {
		t.Error("expected expired challenge to be cleaned up")
	}
	if !exists2 {
		t.Error("expected non-expired challenge to remain")
	}
}
