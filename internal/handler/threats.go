package handler

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ThreatEvent represents a detected attack or suspicious pattern.
type ThreatEvent struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // brute_force, credential_stuffing, bot_wave, rate_abuse, scraping
	Severity    string    `json:"severity"` // low, medium, high, critical
	SourceIP    string    `json:"source_ip"`
	TargetScene string    `json:"target_scene"`
	Details     string    `json:"details"`
	EventCount  int       `json:"event_count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"` // active, mitigated, resolved
}

type ThreatMonitor struct {
	mu     sync.RWMutex
	events []*ThreatEvent
}

func NewThreatMonitor() *ThreatMonitor {
	m := &ThreatMonitor{}
	m.seedEvents()
	return m
}

func (m *ThreatMonitor) seedEvents() {
	now := time.Now()
	m.events = []*ThreatEvent{
		{
			ID: uuid.New().String(), Type: "credential_stuffing", Severity: "critical",
			SourceIP: "185.220.100.x (Tor)", TargetScene: "login",
			Details: "Distributed login attempts across 50+ accounts from Tor exit nodes",
			EventCount: 2847, FirstSeen: now.Add(-2 * time.Hour), LastSeen: now.Add(-15 * time.Minute),
			Status: "active",
		},
		{
			ID: uuid.New().String(), Type: "bot_wave", Severity: "high",
			SourceIP: "159.89.0.0/16 (DigitalOcean)", TargetScene: "register",
			Details: "Registration bot wave from datacenter IPs with HeadlessChrome UA",
			EventCount: 892, FirstSeen: now.Add(-6 * time.Hour), LastSeen: now.Add(-1 * time.Hour),
			Status: "mitigated",
		},
		{
			ID: uuid.New().String(), Type: "rate_abuse", Severity: "medium",
			SourceIP: "203.0.113.50", TargetScene: "api",
			Details: "Single IP exceeding rate limits on API endpoint, 500+ RPM",
			EventCount: 3200, FirstSeen: now.Add(-30 * time.Minute), LastSeen: now.Add(-5 * time.Minute),
			Status: "active",
		},
		{
			ID: uuid.New().String(), Type: "scraping", Severity: "medium",
			SourceIP: "95.216.0.0/15 (Hetzner)", TargetScene: "activity",
			Details: "Systematic crawling pattern with rotating fingerprints",
			EventCount: 456, FirstSeen: now.Add(-12 * time.Hour), LastSeen: now.Add(-3 * time.Hour),
			Status: "resolved",
		},
		{
			ID: uuid.New().String(), Type: "brute_force", Severity: "high",
			SourceIP: "Multiple (CGNAT range)", TargetScene: "login",
			Details: "Password brute force targeting admin accounts",
			EventCount: 1523, FirstSeen: now.Add(-1 * time.Hour), LastSeen: now.Add(-10 * time.Minute),
			Status: "active",
		},
	}
}

func (m *ThreatMonitor) Record(event *ThreatEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append([]*ThreatEvent{event}, m.events...)
	if len(m.events) > 500 {
		m.events = m.events[:500]
	}
}

type ThreatsHandler struct {
	monitor *ThreatMonitor
}

func NewThreatsHandler() *ThreatsHandler {
	return &ThreatsHandler{monitor: NewThreatMonitor()}
}

// List handles GET /v1/threats
func (h *ThreatsHandler) List(c *gin.Context) {
	h.monitor.mu.RLock()
	defer h.monitor.mu.RUnlock()

	status := c.Query("status")
	var filtered []*ThreatEvent
	for _, e := range h.monitor.events {
		if status != "" && e.Status != status {
			continue
		}
		filtered = append(filtered, e)
	}
	if filtered == nil {
		filtered = []*ThreatEvent{}
	}

	c.JSON(http.StatusOK, gin.H{
		"threats": filtered,
		"total":   len(filtered),
		"active":  countByStatus(h.monitor.events, "active"),
	})
}

// Dashboard handles GET /v1/threats/dashboard
func (h *ThreatsHandler) Dashboard(c *gin.Context) {
	h.monitor.mu.RLock()
	events := h.monitor.events
	h.monitor.mu.RUnlock()

	active := countByStatus(events, "active")
	mitigated := countByStatus(events, "mitigated")
	resolved := countByStatus(events, "resolved")

	byType := make(map[string]int)
	bySeverity := make(map[string]int)
	for _, e := range events {
		byType[e.Type]++
		bySeverity[e.Severity]++
	}

	// Attack timeline (last 24h, hourly)
	var timeline []gin.H
	for h := 0; h < 24; h++ {
		timeline = append(timeline, gin.H{
			"hour":       h,
			"attacks":    rand.Intn(50) + 5,
			"blocked":    rand.Intn(40) + 3,
			"challenges": rand.Intn(200) + 50,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"active_threats":    active,
			"mitigated_threats": mitigated,
			"resolved_threats":  resolved,
			"total_threats":     len(events),
		},
		"by_type":     byType,
		"by_severity": bySeverity,
		"timeline":    timeline,
		"top_ips": []gin.H{
			{"ip": "185.220.100.252", "events": 2847, "type": "tor_exit", "blocked": true},
			{"ip": "159.89.42.15", "events": 892, "type": "datacenter", "blocked": true},
			{"ip": "203.0.113.50", "events": 3200, "type": "rate_abuse", "blocked": false},
			{"ip": "95.216.50.1", "events": 456, "type": "scraper", "blocked": true},
			{"ip": "100.64.12.88", "events": 1523, "type": "cgnat", "blocked": false},
		},
	})
}

// Mitigate handles POST /v1/threats/:id/mitigate
func (h *ThreatsHandler) Mitigate(c *gin.Context) {
	id := c.Param("id")
	h.monitor.mu.Lock()
	defer h.monitor.mu.Unlock()

	for _, e := range h.monitor.events {
		if e.ID == id {
			e.Status = "mitigated"
			c.JSON(http.StatusOK, gin.H{"status": "mitigated", "id": id})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "threat not found"})
}

func countByStatus(events []*ThreatEvent, status string) int {
	count := 0
	for _, e := range events {
		if e.Status == status {
			count++
		}
	}
	return count
}
