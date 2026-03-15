package policy

import (
	"testing"

	"github.com/engagelab/captcha/internal/model"
)

func TestAdaptive_HighFalsePositiveRateLowersThreshold(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	scene := model.SceneTypeLogin

	// Simulate outcomes: many humans challenged and failing (false positives).
	for i := 0; i < 25; i++ {
		// Human was challenged and did NOT pass → false positive.
		ae.RecordOutcome(scene, false, true, false)
	}

	// Also record some successful challenges to meet the minimum sample size.
	for i := 0; i < 5; i++ {
		ae.RecordOutcome(scene, false, true, true)
	}

	adj := ae.GetAdjustment(scene)
	if adj <= 0 {
		t.Errorf("expected positive adjustment (raise thresholds) due to high FP rate, got %f", adj)
	}

	// Verify the adapted policy has higher thresholds.
	original := &model.Policy{
		ThresholdLow:  20,
		ThresholdHigh: 60,
		Enabled:       true,
	}
	adapted := ae.GetAdaptedPolicy(scene, original)
	if adapted.ThresholdLow <= original.ThresholdLow {
		t.Errorf("expected adapted ThresholdLow > original: adapted=%f, original=%f",
			adapted.ThresholdLow, original.ThresholdLow)
	}
}

func TestAdaptive_HighBotPassRateRaisesThreshold(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	scene := model.SceneTypeRegister

	// Simulate outcomes: many bots passing challenges.
	for i := 0; i < 25; i++ {
		// Bot was challenged and passed → bot leaking through.
		ae.RecordOutcome(scene, true, true, true)
	}
	// Some bots blocked.
	for i := 0; i < 5; i++ {
		ae.RecordOutcome(scene, true, true, false)
	}

	adj := ae.GetAdjustment(scene)
	if adj >= 0 {
		t.Errorf("expected negative adjustment (lower thresholds) due to high bot pass rate, got %f", adj)
	}

	// Verify adapted policy has lower thresholds.
	original := &model.Policy{
		ThresholdLow:  20,
		ThresholdHigh: 60,
		Enabled:       true,
	}
	adapted := ae.GetAdaptedPolicy(scene, original)
	if adapted.ThresholdLow >= original.ThresholdLow {
		t.Errorf("expected adapted ThresholdLow < original: adapted=%f, original=%f",
			adapted.ThresholdLow, original.ThresholdLow)
	}
}

func TestAdaptive_AdjustmentsCapped(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	scene := model.SceneTypeLogin

	// Push adjustments to the max by repeatedly triggering high FP rates.
	// Each cycle: record a batch of outcomes then check.
	for round := 0; round < 10; round++ {
		// Reset metrics to force re-evaluation.
		ae.mu.Lock()
		ae.scenes[scene] = &sceneMetrics{}
		ae.mu.Unlock()

		for i := 0; i < 25; i++ {
			ae.RecordOutcome(scene, false, true, false)
		}
		for i := 0; i < 5; i++ {
			ae.RecordOutcome(scene, false, true, true)
		}
	}

	adj := ae.GetAdjustment(scene)
	if adj > 20 {
		t.Errorf("expected adjustment capped at +20, got %f", adj)
	}

	// Now do the same for negative direction.
	scene2 := model.SceneTypeRegister
	for round := 0; round < 10; round++ {
		ae.mu.Lock()
		ae.scenes[scene2] = &sceneMetrics{}
		ae.mu.Unlock()

		for i := 0; i < 25; i++ {
			ae.RecordOutcome(scene2, true, true, true)
		}
		for i := 0; i < 5; i++ {
			ae.RecordOutcome(scene2, true, true, false)
		}
	}

	adj2 := ae.GetAdjustment(scene2)
	if adj2 < -20 {
		t.Errorf("expected adjustment capped at -20, got %f", adj2)
	}
}

func TestAdaptive_NormalRatesNoAdjustment(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	scene := model.SceneTypeComment

	// Simulate normal outcomes: low false positive rate, low bot pass rate.
	for i := 0; i < 20; i++ {
		ae.RecordOutcome(scene, false, true, true) // humans pass
	}
	for i := 0; i < 5; i++ {
		ae.RecordOutcome(scene, true, true, false) // bots fail
	}

	adj := ae.GetAdjustment(scene)
	if adj != 0 {
		t.Errorf("expected no adjustment for normal rates, got %f", adj)
	}
}

func TestAdaptive_NilPolicyReturnsNil(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	result := ae.GetAdaptedPolicy(model.SceneTypeLogin, nil)
	if result != nil {
		t.Error("expected nil for nil original policy")
	}
}

func TestAdaptive_BelowMinSampleNoAdjust(t *testing.T) {
	base := NewEngine()
	ae := NewAdaptiveEngine(base)

	scene := model.SceneTypeAPI

	// Only a few outcomes, below the minimum threshold of 20.
	for i := 0; i < 10; i++ {
		ae.RecordOutcome(scene, false, true, false)
	}

	adj := ae.GetAdjustment(scene)
	if adj != 0 {
		t.Errorf("expected no adjustment below minimum sample size, got %f", adj)
	}
}
