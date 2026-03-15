package threat_intel

import (
	"math"
	"sort"
	"sync"
	"time"
)

// ThreatProfile holds aggregated threat intelligence for an IP or fingerprint.
type ThreatProfile struct {
	SeenCount   int       `json:"seen_count"`
	ThreatTypes []string  `json:"threat_types"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	RiskBoost   float64   `json:"risk_boost"`
}

// ThreatReport is an individual anonymized threat signal.
type ThreatReport struct {
	ThreatType string
	Severity   int // 1-10
	ReportedAt time.Time
	// Source identifier (hashed tenant ID) to count unique reporters.
	SourceHash string
}

// threatRecord stores all reports for a single IP or fingerprint.
type threatRecord struct {
	reports     []ThreatReport
	firstSeen   time.Time
	lastSeen    time.Time
	threatTypes map[string]bool
	sources     map[string]bool // unique anonymous reporter hashes
}

// TopThreat represents a highly reported IP or fingerprint.
type TopThreat struct {
	Key         string   `json:"key"`
	SeenCount   int      `json:"seen_count"`
	RiskBoost   float64  `json:"risk_boost"`
	ThreatTypes []string `json:"threat_types"`
}

// ThreatIntelNetwork provides cross-tenant anonymous threat intelligence.
// No tenant identifiers are stored; only hashed source identifiers are used
// to count the number of unique reporters.
type ThreatIntelNetwork struct {
	mu sync.RWMutex

	ipRecords map[string]*threatRecord
	fpRecords map[string]*threatRecord

	// Half-life for auto-decay (default 24h).
	halfLife time.Duration
}

// NewThreatIntelNetwork creates a new cross-tenant threat intelligence network.
func NewThreatIntelNetwork() *ThreatIntelNetwork {
	return &ThreatIntelNetwork{
		ipRecords: make(map[string]*threatRecord),
		fpRecords: make(map[string]*threatRecord),
		halfLife:  24 * time.Hour,
	}
}

// ReportThreat records an anonymized threat signal. The sourceHash is a hash
// of the tenant ID so we can count unique reporters without storing tenant info.
func (n *ThreatIntelNetwork) ReportThreat(ip, fingerprint, threatType string, severity int, sourceHash string) {
	now := time.Now()
	report := ThreatReport{
		ThreatType: threatType,
		Severity:   clampSeverity(severity),
		ReportedAt: now,
		SourceHash: sourceHash,
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if ip != "" {
		rec := n.getOrCreateRecord(n.ipRecords, ip, now)
		rec.reports = append(rec.reports, report)
		rec.lastSeen = now
		rec.threatTypes[threatType] = true
		rec.sources[sourceHash] = true
	}

	if fingerprint != "" {
		rec := n.getOrCreateRecord(n.fpRecords, fingerprint, now)
		rec.reports = append(rec.reports, report)
		rec.lastSeen = now
		rec.threatTypes[threatType] = true
		rec.sources[sourceHash] = true
	}
}

// QueryIP returns the threat profile for a given IP address.
func (n *ThreatIntelNetwork) QueryIP(ip string) ThreatProfile {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.buildProfile(n.ipRecords[ip])
}

// QueryFingerprint returns the threat profile for a given fingerprint.
func (n *ThreatIntelNetwork) QueryFingerprint(fp string) ThreatProfile {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.buildProfile(n.fpRecords[fp])
}

// TopThreats returns the top N most reported IPs.
func (n *ThreatIntelNetwork) TopThreats(limit int) []TopThreat {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var threats []TopThreat
	for key, rec := range n.ipRecords {
		profile := n.buildProfile(rec)
		threats = append(threats, TopThreat{
			Key:         key,
			SeenCount:   profile.SeenCount,
			RiskBoost:   profile.RiskBoost,
			ThreatTypes: profile.ThreatTypes,
		})
	}

	sort.Slice(threats, func(i, j int) bool {
		return threats[i].RiskBoost > threats[j].RiskBoost
	})

	if limit > 0 && len(threats) > limit {
		threats = threats[:limit]
	}
	return threats
}

// buildProfile constructs a ThreatProfile from a record, applying time-decay.
func (n *ThreatIntelNetwork) buildProfile(rec *threatRecord) ThreatProfile {
	if rec == nil {
		return ThreatProfile{}
	}

	now := time.Now()
	var totalDecayedScore float64

	for _, r := range rec.reports {
		age := now.Sub(r.ReportedAt)
		// Exponential decay: score * 2^(-age/halfLife)
		decayFactor := math.Pow(2, -float64(age)/float64(n.halfLife))
		totalDecayedScore += float64(r.Severity) * decayFactor
	}

	// Cross-tenant multiplier: more unique sources = higher risk.
	sourceCount := len(rec.sources)
	crossTenantMultiplier := 1.0 + float64(sourceCount-1)*0.25
	if crossTenantMultiplier < 1.0 {
		crossTenantMultiplier = 1.0
	}

	riskBoost := totalDecayedScore * crossTenantMultiplier
	// Cap risk boost at 50 to prevent unbounded growth.
	if riskBoost > 50 {
		riskBoost = 50
	}

	types := make([]string, 0, len(rec.threatTypes))
	for t := range rec.threatTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	return ThreatProfile{
		SeenCount:   len(rec.reports),
		ThreatTypes: types,
		FirstSeen:   rec.firstSeen,
		LastSeen:    rec.lastSeen,
		RiskBoost:   math.Round(riskBoost*100) / 100,
	}
}

func (n *ThreatIntelNetwork) getOrCreateRecord(records map[string]*threatRecord, key string, now time.Time) *threatRecord {
	rec, ok := records[key]
	if !ok {
		rec = &threatRecord{
			firstSeen:   now,
			lastSeen:    now,
			threatTypes: make(map[string]bool),
			sources:     make(map[string]bool),
		}
		records[key] = rec
	}
	return rec
}

func clampSeverity(s int) int {
	if s < 1 {
		return 1
	}
	if s > 10 {
		return 10
	}
	return s
}
