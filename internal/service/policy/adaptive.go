package policy

import (
	"log"
	"sync"

	"github.com/engagelab/captcha/internal/model"
)

// sceneMetrics tracks per-scene outcome metrics for adaptive tuning.
type sceneMetrics struct {
	totalChallenges int
	totalPassed     int
	botsChallenged  int
	botsPassed      int
	humansChallenged int
	humansFailed    int // false positives: humans that were challenged or denied
}

// AdaptiveEngine wraps the base policy engine and auto-adjusts thresholds
// based on observed false positive and bot pass rates.
type AdaptiveEngine struct {
	mu     sync.Mutex
	base   *Engine
	scenes map[model.SceneType]*sceneMetrics
	// adjustments stores the current threshold delta per scene.
	adjustments map[model.SceneType]float64
	// maxAdjustment caps how far thresholds can shift from their original values.
	maxAdjustment float64
}

// NewAdaptiveEngine creates an adaptive engine wrapping the base policy engine.
func NewAdaptiveEngine(base *Engine) *AdaptiveEngine {
	return &AdaptiveEngine{
		base:          base,
		scenes:        make(map[model.SceneType]*sceneMetrics),
		adjustments:   make(map[model.SceneType]float64),
		maxAdjustment: 20,
	}
}

// RecordOutcome records the outcome of a challenge for adaptive tuning.
//   - scene: the scene type
//   - wasBot: whether the request was ultimately determined to be a bot
//   - wasChallenged: whether the system issued a challenge (or denied)
//   - passed: whether the challenge was passed
func (ae *AdaptiveEngine) RecordOutcome(scene model.SceneType, wasBot, wasChallenged, passed bool) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	m := ae.getOrCreateMetrics(scene)

	if wasChallenged {
		m.totalChallenges++
		if passed {
			m.totalPassed++
		}
		if wasBot {
			m.botsChallenged++
			if passed {
				m.botsPassed++
			}
		} else {
			m.humansChallenged++
			if !passed {
				m.humansFailed++
			}
		}
	} else if wasBot && !wasChallenged {
		// Bot that was not challenged at all counts as bot passed.
		m.botsPassed++
		m.botsChallenged++ // Count for rate calculation.
	}

	ae.maybeAdjust(scene)
}

// GetAdaptedPolicy returns a copy of the policy with adjusted thresholds.
func (ae *AdaptiveEngine) GetAdaptedPolicy(sceneType model.SceneType, original *model.Policy) *model.Policy {
	if original == nil {
		return nil
	}

	ae.mu.Lock()
	adj := ae.adjustments[sceneType]
	ae.mu.Unlock()

	if adj == 0 {
		return original
	}

	// Create a copy with adjusted thresholds.
	adapted := *original
	adapted.ThresholdLow += adj
	adapted.ThresholdHigh += adj

	// Clamp thresholds to valid range.
	if adapted.ThresholdLow < 0 {
		adapted.ThresholdLow = 0
	}
	if adapted.ThresholdHigh < adapted.ThresholdLow+1 {
		adapted.ThresholdHigh = adapted.ThresholdLow + 1
	}
	if adapted.ThresholdHigh > 100 {
		adapted.ThresholdHigh = 100
	}

	return &adapted
}

// GetAdjustment returns the current threshold adjustment for a scene type.
func (ae *AdaptiveEngine) GetAdjustment(sceneType model.SceneType) float64 {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	return ae.adjustments[sceneType]
}

// GetMetricsSnapshot returns the current metrics for a scene type.
func (ae *AdaptiveEngine) GetMetricsSnapshot(sceneType model.SceneType) (falsePositiveRate, botPassRate float64) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	m, ok := ae.scenes[sceneType]
	if !ok {
		return 0, 0
	}
	return ae.calcFalsePositiveRate(m), ae.calcBotPassRate(m)
}

// maybeAdjust checks rates and adjusts thresholds if needed.
// Must be called with ae.mu held.
func (ae *AdaptiveEngine) maybeAdjust(scene model.SceneType) {
	m := ae.scenes[scene]
	if m == nil {
		return
	}

	// Need a minimum sample size before adjusting.
	if m.totalChallenges < 20 {
		return
	}

	fpr := ae.calcFalsePositiveRate(m)
	bpr := ae.calcBotPassRate(m)

	currentAdj := ae.adjustments[scene]

	// If false positive rate > 5%, lower thresholds (raise them numerically so fewer
	// requests trigger challenges). This means adding a positive adjustment.
	if fpr > 5.0 {
		newAdj := currentAdj + 5
		if newAdj > ae.maxAdjustment {
			newAdj = ae.maxAdjustment
		}
		if newAdj != currentAdj {
			ae.adjustments[scene] = newAdj
			log.Printf("[adaptive] scene=%s: false_positive_rate=%.1f%% > 5%%, raising thresholds by 5 (adj=%.0f)",
				scene, fpr, newAdj)
		}
	}

	// If bot pass rate > 10%, raise thresholds (lower them numerically so more
	// requests trigger challenges). This means subtracting from the adjustment.
	if bpr > 10.0 {
		newAdj := currentAdj - 5
		if newAdj < -ae.maxAdjustment {
			newAdj = -ae.maxAdjustment
		}
		if newAdj != currentAdj {
			ae.adjustments[scene] = newAdj
			log.Printf("[adaptive] scene=%s: bot_pass_rate=%.1f%% > 10%%, lowering thresholds by 5 (adj=%.0f)",
				scene, bpr, newAdj)
		}
	}
}

func (ae *AdaptiveEngine) calcFalsePositiveRate(m *sceneMetrics) float64 {
	if m.humansChallenged == 0 {
		return 0
	}
	return float64(m.humansFailed) / float64(m.humansChallenged) * 100
}

func (ae *AdaptiveEngine) calcBotPassRate(m *sceneMetrics) float64 {
	if m.botsChallenged == 0 {
		return 0
	}
	return float64(m.botsPassed) / float64(m.botsChallenged) * 100
}

func (ae *AdaptiveEngine) getOrCreateMetrics(scene model.SceneType) *sceneMetrics {
	m, ok := ae.scenes[scene]
	if !ok {
		m = &sceneMetrics{}
		ae.scenes[scene] = m
	}
	return m
}
