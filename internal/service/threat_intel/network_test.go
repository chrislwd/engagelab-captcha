package threat_intel

import (
	"math"
	"testing"
	"time"
)

func TestReportAndQueryIP(t *testing.T) {
	n := NewThreatIntelNetwork()
	n.ReportThreat("1.2.3.4", "", "brute_force", 5, "tenant-a")

	profile := n.QueryIP("1.2.3.4")
	if profile.SeenCount != 1 {
		t.Errorf("expected seen_count=1, got %d", profile.SeenCount)
	}
	if len(profile.ThreatTypes) != 1 || profile.ThreatTypes[0] != "brute_force" {
		t.Errorf("expected threat_types=[brute_force], got %v", profile.ThreatTypes)
	}
	if profile.RiskBoost <= 0 {
		t.Errorf("expected positive risk_boost, got %f", profile.RiskBoost)
	}
}

func TestRiskBoostIncreasesWithReports(t *testing.T) {
	n := NewThreatIntelNetwork()

	n.ReportThreat("1.2.3.4", "", "bot", 5, "tenant-a")
	boost1 := n.QueryIP("1.2.3.4").RiskBoost

	n.ReportThreat("1.2.3.4", "", "scraping", 7, "tenant-a")
	boost2 := n.QueryIP("1.2.3.4").RiskBoost

	if boost2 <= boost1 {
		t.Errorf("expected risk_boost to increase with more reports: %f -> %f", boost1, boost2)
	}
}

func TestCrossTenantReportsIncreaseScore(t *testing.T) {
	n := NewThreatIntelNetwork()

	// Single tenant reporting.
	n.ReportThreat("5.5.5.5", "", "bot", 5, "tenant-a")
	singleTenantBoost := n.QueryIP("5.5.5.5").RiskBoost

	// Different tenant reporting the same IP.
	n.ReportThreat("5.5.5.5", "", "bot", 5, "tenant-b")
	multiTenantBoost := n.QueryIP("5.5.5.5").RiskBoost

	if multiTenantBoost <= singleTenantBoost {
		t.Errorf("cross-tenant reports should increase score: single=%f, multi=%f",
			singleTenantBoost, multiTenantBoost)
	}
}

func TestAutoDecayReducesScore(t *testing.T) {
	n := NewThreatIntelNetwork()
	// Use a very short half-life for testing.
	n.halfLife = 1 * time.Millisecond

	n.ReportThreat("9.8.7.6", "", "bot", 10, "tenant-a")
	boostBefore := n.QueryIP("9.8.7.6").RiskBoost

	// Wait for several half-lives.
	time.Sleep(50 * time.Millisecond)

	boostAfter := n.QueryIP("9.8.7.6").RiskBoost

	if boostAfter >= boostBefore {
		t.Errorf("expected decayed score to be lower: before=%f, after=%f",
			boostBefore, boostAfter)
	}

	// After many half-lives, score should be near zero.
	if boostAfter > 0.1 {
		t.Errorf("expected score near zero after many half-lives, got %f", boostAfter)
	}
}

func TestTopThreatsListing(t *testing.T) {
	n := NewThreatIntelNetwork()

	n.ReportThreat("10.0.0.1", "", "bot", 3, "t1")
	n.ReportThreat("10.0.0.2", "", "brute_force", 8, "t1")
	n.ReportThreat("10.0.0.3", "", "scraping", 5, "t1")
	// Report 10.0.0.2 again from another tenant to boost its score.
	n.ReportThreat("10.0.0.2", "", "scraping", 7, "t2")

	top := n.TopThreats(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 top threats, got %d", len(top))
	}
	// 10.0.0.2 should be first (highest risk_boost).
	if top[0].Key != "10.0.0.2" {
		t.Errorf("expected top threat to be 10.0.0.2, got %s", top[0].Key)
	}
}

func TestQueryUnknownIPReturnsEmpty(t *testing.T) {
	n := NewThreatIntelNetwork()
	profile := n.QueryIP("99.99.99.99")
	if profile.SeenCount != 0 {
		t.Errorf("expected seen_count=0 for unknown IP, got %d", profile.SeenCount)
	}
	if profile.RiskBoost != 0 {
		t.Errorf("expected risk_boost=0 for unknown IP, got %f", profile.RiskBoost)
	}
}

func TestQueryFingerprint(t *testing.T) {
	n := NewThreatIntelNetwork()
	n.ReportThreat("", "fp-evil", "credential_stuffing", 9, "tenant-x")

	profile := n.QueryFingerprint("fp-evil")
	if profile.SeenCount != 1 {
		t.Errorf("expected seen_count=1, got %d", profile.SeenCount)
	}
	if profile.RiskBoost <= 0 {
		t.Errorf("expected positive risk_boost, got %f", profile.RiskBoost)
	}
}

func TestSeverityClamped(t *testing.T) {
	n := NewThreatIntelNetwork()
	n.ReportThreat("1.1.1.1", "", "test", 100, "t1") // should clamp to 10
	profile := n.QueryIP("1.1.1.1")

	// With severity clamped to 10, risk_boost should not exceed 10 (single report, single source).
	if profile.RiskBoost > 10.01 {
		t.Errorf("expected clamped severity to limit boost, got %f", profile.RiskBoost)
	}
}

func TestRiskBoostCapped(t *testing.T) {
	n := NewThreatIntelNetwork()

	// Many reports to try to exceed the cap.
	for i := 0; i < 100; i++ {
		n.ReportThreat("2.2.2.2", "", "flood", 10, "tenant-flood")
	}

	profile := n.QueryIP("2.2.2.2")
	if profile.RiskBoost > 50 {
		t.Errorf("expected risk_boost capped at 50, got %f", profile.RiskBoost)
	}
}

func TestDecayFormula(t *testing.T) {
	// Verify the decay formula directly: after exactly 1 half-life, score should be halved.
	n := NewThreatIntelNetwork()
	n.halfLife = 100 * time.Millisecond

	n.mu.Lock()
	rec := n.getOrCreateRecord(n.ipRecords, "decay-test", time.Now().Add(-100*time.Millisecond))
	rec.reports = append(rec.reports, ThreatReport{
		ThreatType: "test",
		Severity:   10,
		ReportedAt: time.Now().Add(-100 * time.Millisecond),
		SourceHash: "s1",
	})
	rec.sources["s1"] = true
	rec.threatTypes["test"] = true
	n.mu.Unlock()

	profile := n.QueryIP("decay-test")
	// After 1 half-life, decayed score should be ~5 (10 * 0.5).
	expected := 5.0
	if math.Abs(profile.RiskBoost-expected) > 1.0 {
		t.Errorf("expected risk_boost ~%f after 1 half-life, got %f", expected, profile.RiskBoost)
	}
}
