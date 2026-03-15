package challenge

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"time"
)

// ABTest represents an active experiment comparing challenge configurations.
type ABTest struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SceneID     string    `json:"scene_id"`
	Status      string    `json:"status"` // active, paused, completed
	Variants    []Variant `json:"variants"`
	TrafficPct  int       `json:"traffic_pct"` // percentage of traffic in experiment
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Variant is a single arm of an A/B test.
type Variant struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ChallengeType  string `json:"challenge_type"`
	Weight         int    `json:"weight"` // relative weight for traffic split
	Impressions    int64  `json:"impressions"`
	Completions    int64  `json:"completions"`
	Failures       int64  `json:"failures"`
	AvgDurationMs  int64  `json:"avg_duration_ms"`
	ConversionRate float64 `json:"conversion_rate"`
}

// ABTestManager manages challenge A/B experiments.
type ABTestManager struct {
	mu    sync.RWMutex
	tests map[string]*ABTest // scene_id -> active test
}

func NewABTestManager() *ABTestManager {
	return &ABTestManager{
		tests: make(map[string]*ABTest),
	}
}

// CreateTest starts a new A/B test for a scene.
func (m *ABTestManager) CreateTest(test *ABTest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	test.Status = "active"
	test.StartedAt = time.Now()
	m.tests[test.SceneID] = test
}

// GetVariant assigns a user to a variant using deterministic hashing.
// The same fingerprint always gets the same variant for consistency.
func (m *ABTestManager) GetVariant(sceneID, fingerprint string) *Variant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	test, ok := m.tests[sceneID]
	if !ok || test.Status != "active" {
		return nil
	}

	// Deterministic assignment based on fingerprint hash
	h := sha256.Sum256([]byte(fingerprint + test.ID))
	bucket := int(binary.BigEndian.Uint32(h[:4])) % 100

	// Check if this request is in the experiment traffic
	if bucket >= test.TrafficPct {
		return nil // Not in experiment
	}

	// Weighted variant selection
	totalWeight := 0
	for _, v := range test.Variants {
		totalWeight += v.Weight
	}
	if totalWeight == 0 {
		return nil
	}

	variantBucket := int(binary.BigEndian.Uint32(h[4:8])) % totalWeight
	cumulative := 0
	for i := range test.Variants {
		cumulative += test.Variants[i].Weight
		if variantBucket < cumulative {
			return &test.Variants[i]
		}
	}

	return &test.Variants[0]
}

// RecordResult updates variant statistics.
func (m *ABTestManager) RecordResult(sceneID, variantID string, passed bool, durationMs int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	test, ok := m.tests[sceneID]
	if !ok {
		return
	}

	for i := range test.Variants {
		if test.Variants[i].ID == variantID {
			test.Variants[i].Impressions++
			if passed {
				test.Variants[i].Completions++
			} else {
				test.Variants[i].Failures++
			}
			// Update rolling average
			total := test.Variants[i].Impressions
			test.Variants[i].AvgDurationMs = (test.Variants[i].AvgDurationMs*(total-1) + durationMs) / total
			if test.Variants[i].Impressions > 0 {
				test.Variants[i].ConversionRate = float64(test.Variants[i].Completions) / float64(test.Variants[i].Impressions) * 100
			}
			break
		}
	}
}

// GetResults returns current test results for a scene.
func (m *ABTestManager) GetResults(sceneID string) *ABTest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	test, ok := m.tests[sceneID]
	if !ok {
		return nil
	}
	// Return a copy
	copy := *test
	return &copy
}

// ListTests returns all tests.
func (m *ABTestManager) ListTests() []*ABTest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*ABTest
	for _, t := range m.tests {
		copy := *t
		result = append(result, &copy)
	}
	return result
}
