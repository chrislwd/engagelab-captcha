package sms

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SMSCheckResult holds the outcome of an SMS rate/abuse check.
type SMSCheckResult struct {
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason"`
	WaitSeconds int    `json:"wait_seconds"`
}

// tenantCost tracks daily SMS spend per tenant.
type tenantCost struct {
	Date      string // YYYY-MM-DD
	CostCents int
}

// SMSProtector enforces rate limits and abuse prevention for SMS sending.
type SMSProtector struct {
	mu sync.Mutex

	// Per-phone request timestamps.
	phoneCounters map[string][]time.Time
	// Per-IP request timestamps.
	ipCounters map[string][]time.Time
	// Per-fingerprint request timestamps.
	fpCounters map[string][]time.Time

	// Daily SMS cost tracking per tenant (keyed by app_id).
	tenantCosts map[string]*tenantCost

	// Cost per SMS in cents.
	costPerSMSCents int
}

// NewSMSProtector creates a new SMS abuse prevention service.
func NewSMSProtector() *SMSProtector {
	return &SMSProtector{
		phoneCounters:   make(map[string][]time.Time),
		ipCounters:      make(map[string][]time.Time),
		fpCounters:      make(map[string][]time.Time),
		tenantCosts:     make(map[string]*tenantCost),
		costPerSMSCents: 1, // 1 cent per SMS
	}
}

// e164Pattern matches E.164 phone numbers: + followed by 1-15 digits.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// disposablePrefixes contains known disposable/virtual number prefixes.
var disposablePrefixes = []string{
	"+1900",  // US premium rate
	"+44700", // UK personal numbering
	"+44870", // UK non-geographic
	"+44871", // UK non-geographic
	"+33891", // France premium
	"+1555",  // US fictional
}

// CheckSMSRequest determines whether an SMS should be allowed based on rate
// limits, phone format validation, and disposable number detection.
func (p *SMSProtector) CheckSMSRequest(phoneNumber, ip, fingerprint string) SMSCheckResult {
	// 1. Validate phone format (E.164).
	if !e164Pattern.MatchString(phoneNumber) {
		return SMSCheckResult{
			Allowed:     false,
			Reason:      "invalid phone number format, must be E.164 (e.g. +14155552671)",
			WaitSeconds: 0,
		}
	}

	// 2. Check for disposable/virtual numbers.
	for _, prefix := range disposablePrefixes {
		if strings.HasPrefix(phoneNumber, prefix) {
			return SMSCheckResult{
				Allowed:     false,
				Reason:      "disposable or virtual phone number not allowed",
				WaitSeconds: 0,
			}
		}
	}

	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	// 3. Per-phone rate limits.
	p.phoneCounters[phoneNumber] = pruneOld(p.phoneCounters[phoneNumber], now, time.Hour)
	phoneTimes := p.phoneCounters[phoneNumber]

	// Rule: max 1 SMS per phone per 60 seconds.
	if len(phoneTimes) > 0 {
		lastSent := phoneTimes[len(phoneTimes)-1]
		elapsed := now.Sub(lastSent)
		if elapsed < 60*time.Second {
			wait := int((60*time.Second - elapsed).Seconds()) + 1
			return SMSCheckResult{
				Allowed:     false,
				Reason:      "too soon, please wait before requesting another SMS",
				WaitSeconds: wait,
			}
		}
	}

	// Rule: max 5 SMS per phone per hour.
	if len(phoneTimes) >= 5 {
		oldest := phoneTimes[len(phoneTimes)-5]
		wait := int(time.Hour.Seconds() - now.Sub(oldest).Seconds()) + 1
		if wait < 1 {
			wait = 1
		}
		return SMSCheckResult{
			Allowed:     false,
			Reason:      "phone number hourly limit exceeded",
			WaitSeconds: wait,
		}
	}

	// 4. Per-IP rate limits: max 10 per IP per hour.
	if ip != "" {
		p.ipCounters[ip] = pruneOld(p.ipCounters[ip], now, time.Hour)
		if len(p.ipCounters[ip]) >= 10 {
			oldest := p.ipCounters[ip][len(p.ipCounters[ip])-10]
			wait := int(time.Hour.Seconds()-now.Sub(oldest).Seconds()) + 1
			if wait < 1 {
				wait = 1
			}
			return SMSCheckResult{
				Allowed:     false,
				Reason:      "IP hourly limit exceeded",
				WaitSeconds: wait,
			}
		}
	}

	// 5. Per-fingerprint rate limits: max 20 per fingerprint per hour.
	if fingerprint != "" {
		p.fpCounters[fingerprint] = pruneOld(p.fpCounters[fingerprint], now, time.Hour)
		if len(p.fpCounters[fingerprint]) >= 20 {
			oldest := p.fpCounters[fingerprint][len(p.fpCounters[fingerprint])-20]
			wait := int(time.Hour.Seconds()-now.Sub(oldest).Seconds()) + 1
			if wait < 1 {
				wait = 1
			}
			return SMSCheckResult{
				Allowed:     false,
				Reason:      "fingerprint hourly limit exceeded",
				WaitSeconds: wait,
			}
		}
	}

	// All checks passed — record this request.
	p.phoneCounters[phoneNumber] = append(p.phoneCounters[phoneNumber], now)
	if ip != "" {
		p.ipCounters[ip] = append(p.ipCounters[ip], now)
	}
	if fingerprint != "" {
		p.fpCounters[fingerprint] = append(p.fpCounters[fingerprint], now)
	}

	return SMSCheckResult{
		Allowed:     true,
		Reason:      "allowed",
		WaitSeconds: 0,
	}
}

// RecordSMSCost increments the daily SMS cost for a tenant.
func (p *SMSProtector) RecordSMSCost(appID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	tc, ok := p.tenantCosts[appID]
	if !ok || tc.Date != today {
		p.tenantCosts[appID] = &tenantCost{Date: today, CostCents: p.costPerSMSCents}
		return
	}
	tc.CostCents += p.costPerSMSCents
}

// GetDailyCostCents returns the current day's SMS cost in cents for a tenant.
func (p *SMSProtector) GetDailyCostCents(appID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	tc, ok := p.tenantCosts[appID]
	if !ok || tc.Date != today {
		return 0
	}
	return tc.CostCents
}

// pruneOld removes timestamps older than the given window.
func pruneOld(timestamps []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	start := 0
	for start < len(timestamps) && timestamps[start].Before(cutoff) {
		start++
	}
	return timestamps[start:]
}

// FormatE164 is a helper that attempts to format a raw phone number to E.164.
// It returns the formatted number and an error if the input is invalid.
func FormatE164(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty phone number")
	}
	// Already E.164.
	if e164Pattern.MatchString(raw) {
		return raw, nil
	}
	return "", fmt.Errorf("phone number %q is not in E.164 format", raw)
}
