package risk

import (
	"crypto/sha256"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/engagelab/captcha/internal/model"
)

// Result holds the output of a risk assessment.
type Result struct {
	Score  float64
	Label  string
	Action model.RiskAction
}

// Engine performs risk scoring based on IP, fingerprint, behavior, and rate signals.
type Engine struct {
	mu sync.RWMutex

	// Per-IP request counters: key = IP, value = list of timestamps
	ipCounters map[string][]time.Time
	// Per-fingerprint request counters
	fpCounters map[string][]time.Time
	// Known bad IPs (simulated threat intel)
	badIPs map[string]bool

	// Datacenter CIDR ranges (simplified list of well-known cloud provider ranges)
	datacenterCIDRs []*net.IPNet

	// Enhanced detectors
	proxyDetector *ProxyDetector
	botDetector   *BotPatternDetector
}

// NewEngine creates a new risk assessment engine.
func NewEngine() *Engine {
	e := &Engine{
		ipCounters: make(map[string][]time.Time),
		fpCounters: make(map[string][]time.Time),
		badIPs:     make(map[string]bool),
	}

	// Populate known datacenter CIDRs (subset of major cloud providers).
	cidrs := []string{
		"35.192.0.0/11",   // GCP
		"34.64.0.0/10",    // GCP
		"13.64.0.0/11",    // Azure
		"52.0.0.0/11",     // AWS
		"54.64.0.0/11",    // AWS
		"18.128.0.0/9",    // AWS
		"104.196.0.0/14",  // GCP
		"198.51.100.0/24", // Documentation/test range
	}
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err == nil {
			e.datacenterCIDRs = append(e.datacenterCIDRs, ipnet)
		}
	}

	// Seed some known bad IPs.
	for _, ip := range []string{"192.0.2.1", "198.51.100.50", "203.0.113.99"} {
		e.badIPs[ip] = true
	}

	// Initialize enhanced detectors
	e.proxyDetector = NewProxyDetector()
	e.botDetector = NewBotPatternDetector()

	return e
}

// Evaluate runs a full risk assessment and returns a score from 0-100.
func (e *Engine) Evaluate(ip, userAgent, fingerprint string, behaviorData map[string]interface{}) Result {
	var totalScore float64
	var labels []string

	// 1. IP reputation check (0-30 points)
	ipScore, ipLabels := e.evaluateIP(ip)
	totalScore += ipScore
	labels = append(labels, ipLabels...)

	// 2. Rate limiting signals (0-30 points)
	rateScore, rateLabels := e.evaluateRates(ip, fingerprint)
	totalScore += rateScore
	labels = append(labels, rateLabels...)

	// 3. User-Agent analysis (0-15 points)
	uaScore, uaLabels := e.evaluateUserAgent(userAgent)
	totalScore += uaScore
	labels = append(labels, uaLabels...)

	// 4. Behavior analysis (0-25 points)
	behScore, behLabels := e.evaluateBehavior(behaviorData)
	totalScore += behScore
	labels = append(labels, behLabels...)

	// 5. Enhanced proxy/VPN/Tor detection
	proxyResult := e.proxyDetector.Check(ip)
	totalScore += proxyResult.Score
	labels = append(labels, proxyResult.Labels...)

	// 6. Advanced bot pattern detection
	botResult := e.botDetector.Analyze(userAgent, behaviorData)
	totalScore += botResult.Score
	labels = append(labels, botResult.Labels...)

	// Clamp to 0-100.
	if totalScore > 100 {
		totalScore = 100
	}
	if totalScore < 0 {
		totalScore = 0
	}

	label := classifyRisk(totalScore)
	action := recommendAction(totalScore)

	if len(labels) == 0 {
		labels = append(labels, "clean")
	}

	return Result{
		Score:  math.Round(totalScore*100) / 100,
		Label:  label + "|" + strings.Join(labels, ","),
		Action: action,
	}
}

// evaluateIP checks IP reputation, reserved ranges, and datacenter CIDRs.
func (e *Engine) evaluateIP(ipStr string) (float64, []string) {
	var score float64
	var labels []string

	if ipStr == "" {
		return 10, []string{"missing_ip"}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 15, []string{"invalid_ip"}
	}

	// Check known bad IPs.
	e.mu.RLock()
	if e.badIPs[ipStr] {
		score += 30
		labels = append(labels, "blacklisted_ip")
	}
	e.mu.RUnlock()

	// Check if it is a private/reserved range (could be proxy or VPN).
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		// Private IPs are fine in dev; minor signal.
		score += 2
	}

	// Check datacenter CIDRs.
	for _, cidr := range e.datacenterCIDRs {
		if cidr.Contains(ip) {
			score += 15
			labels = append(labels, "datacenter_ip")
			break
		}
	}

	return score, labels
}

// evaluateRates checks per-IP and per-fingerprint request frequency.
func (e *Engine) evaluateRates(ip, fingerprint string) (float64, []string) {
	var score float64
	var labels []string
	now := time.Now()

	e.mu.Lock()
	defer e.mu.Unlock()

	// Record this request for the IP.
	if ip != "" {
		e.ipCounters[ip] = append(e.ipCounters[ip], now)
		e.ipCounters[ip] = pruneOld(e.ipCounters[ip], now, time.Hour)

		oneMinAgo := now.Add(-time.Minute)
		recentCount := countSince(e.ipCounters[ip], oneMinAgo)

		if recentCount > 30 {
			score += 30
			labels = append(labels, "ip_rate_extreme")
		} else if recentCount > 15 {
			score += 20
			labels = append(labels, "ip_rate_high")
		} else if recentCount > 5 {
			score += 8
			labels = append(labels, "ip_rate_elevated")
		}
	}

	// Record this request for the fingerprint.
	if fingerprint != "" {
		fpKey := hashKey(fingerprint)
		e.fpCounters[fpKey] = append(e.fpCounters[fpKey], now)
		e.fpCounters[fpKey] = pruneOld(e.fpCounters[fpKey], now, time.Hour)

		oneMinAgo := now.Add(-time.Minute)
		recentCount := countSince(e.fpCounters[fpKey], oneMinAgo)

		if recentCount > 20 {
			score += 25
			labels = append(labels, "fp_rate_extreme")
		} else if recentCount > 10 {
			score += 15
			labels = append(labels, "fp_rate_high")
		} else if recentCount > 3 {
			score += 5
			labels = append(labels, "fp_rate_elevated")
		}
	} else {
		// Missing fingerprint is suspicious.
		score += 5
		labels = append(labels, "missing_fingerprint")
	}

	return score, labels
}

// evaluateUserAgent checks for missing, empty, or bot-like User-Agent strings.
func (e *Engine) evaluateUserAgent(ua string) (float64, []string) {
	var score float64
	var labels []string

	if ua == "" {
		return 12, []string{"missing_ua"}
	}

	lower := strings.ToLower(ua)

	// Known bot indicators.
	botTokens := []string{"bot", "crawler", "spider", "headless", "phantom", "selenium", "puppeteer", "playwright", "wget", "curl", "python-requests", "go-http-client", "java/", "libwww"}
	for _, token := range botTokens {
		if strings.Contains(lower, token) {
			score += 15
			labels = append(labels, "bot_ua")
			break
		}
	}

	// Very short UA is suspicious.
	if len(ua) < 20 && score == 0 {
		score += 5
		labels = append(labels, "short_ua")
	}

	return score, labels
}

// evaluateBehavior analyzes client-side behavior signals.
func (e *Engine) evaluateBehavior(data map[string]interface{}) (float64, []string) {
	var score float64
	var labels []string

	if data == nil || len(data) == 0 {
		return 10, []string{"no_behavior_data"}
	}

	// Mouse movement entropy: expect a float64 representing Shannon entropy of movement angles.
	if entropy, ok := getFloat(data, "mouse_entropy"); ok {
		if entropy < 0.5 {
			// Very low entropy indicates scripted, linear movement.
			score += 15
			labels = append(labels, "low_mouse_entropy")
		} else if entropy < 1.5 {
			score += 5
			labels = append(labels, "moderate_mouse_entropy")
		}
		// High entropy (> 1.5) is normal human-like movement.
	}

	// Timing: how long the user spent on the page before triggering CAPTCHA (milliseconds).
	if timing, ok := getFloat(data, "time_on_page_ms"); ok {
		if timing < 500 {
			// Less than 500ms is almost certainly automated.
			score += 15
			labels = append(labels, "instant_interaction")
		} else if timing < 2000 {
			score += 5
			labels = append(labels, "fast_interaction")
		}
	}

	// Key press count: if the form has fields, we expect some key presses.
	if keyCount, ok := getFloat(data, "key_count"); ok {
		if keyCount == 0 {
			score += 5
			labels = append(labels, "no_keystrokes")
		}
	}

	// Scroll events.
	if scrollCount, ok := getFloat(data, "scroll_count"); ok {
		if scrollCount == 0 {
			score += 3
			labels = append(labels, "no_scroll")
		}
	}

	// Touch events on mobile (absence on a mobile UA could indicate emulation).
	if touchEvents, ok := getFloat(data, "touch_count"); ok {
		_ = touchEvents // Presence is enough to reduce suspicion.
	}

	// Presence of devtools signal.
	if devtools, ok := data["devtools_open"].(bool); ok && devtools {
		score += 8
		labels = append(labels, "devtools_open")
	}

	return score, labels
}

// classifyRisk maps a numeric score to a human-readable label.
func classifyRisk(score float64) string {
	switch {
	case score <= 15:
		return "low"
	case score <= 40:
		return "medium"
	case score <= 70:
		return "high"
	default:
		return "critical"
	}
}

// recommendAction maps a numeric score to a default action.
func recommendAction(score float64) model.RiskAction {
	switch {
	case score <= 15:
		return model.RiskActionPass
	case score <= 30:
		return model.RiskActionInvisible
	case score <= 70:
		return model.RiskActionChallenge
	default:
		return model.RiskActionDeny
	}
}

// --- helpers ---

func pruneOld(timestamps []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	start := 0
	for start < len(timestamps) && timestamps[start].Before(cutoff) {
		start++
	}
	return timestamps[start:]
}

func countSince(timestamps []time.Time, since time.Time) int {
	count := 0
	for _, ts := range timestamps {
		if ts.After(since) || ts.Equal(since) {
			count++
		}
	}
	return count
}

func hashKey(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

func getFloat(m map[string]interface{}, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
