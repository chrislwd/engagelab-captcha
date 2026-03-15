package sms

import (
	"testing"
	"time"
)

func TestCheckSMSRequest_NormalAllowed(t *testing.T) {
	p := NewSMSProtector()
	result := p.CheckSMSRequest("+14155552671", "1.2.3.4", "fp-abc")
	if !result.Allowed {
		t.Errorf("expected allowed, got blocked: %s", result.Reason)
	}
}

func TestCheckSMSRequest_SamePhoneWithin60s(t *testing.T) {
	p := NewSMSProtector()

	// First request: should be allowed.
	r1 := p.CheckSMSRequest("+14155552671", "1.2.3.4", "fp-abc")
	if !r1.Allowed {
		t.Fatalf("first request should be allowed, got: %s", r1.Reason)
	}

	// Second request immediately: should be blocked.
	r2 := p.CheckSMSRequest("+14155552671", "1.2.3.4", "fp-abc")
	if r2.Allowed {
		t.Error("second request within 60s should be blocked")
	}
	if r2.WaitSeconds <= 0 {
		t.Error("expected positive wait_seconds")
	}
}

func TestCheckSMSRequest_PhoneHourlyLimitExceeded(t *testing.T) {
	p := NewSMSProtector()
	phone := "+14155552671"

	// Directly inject 5 timestamps in the past (all within the last hour but > 60s apart).
	now := time.Now()
	p.mu.Lock()
	p.phoneCounters[phone] = []time.Time{
		now.Add(-50 * time.Minute),
		now.Add(-40 * time.Minute),
		now.Add(-30 * time.Minute),
		now.Add(-20 * time.Minute),
		now.Add(-2 * time.Minute),
	}
	p.mu.Unlock()

	result := p.CheckSMSRequest(phone, "1.2.3.4", "fp-abc")
	if result.Allowed {
		t.Error("expected phone hourly limit to be exceeded")
	}
	if result.Reason != "phone number hourly limit exceeded" {
		t.Errorf("unexpected reason: %s", result.Reason)
	}
}

func TestCheckSMSRequest_IPHourlyLimitExceeded(t *testing.T) {
	p := NewSMSProtector()
	ip := "1.2.3.4"

	// Inject 10 IP timestamps (different phones, same IP).
	now := time.Now()
	p.mu.Lock()
	for i := 0; i < 10; i++ {
		p.ipCounters[ip] = append(p.ipCounters[ip], now.Add(-time.Duration(50-i*4)*time.Minute))
	}
	p.mu.Unlock()

	result := p.CheckSMSRequest("+14155559999", ip, "fp-xyz")
	if result.Allowed {
		t.Error("expected IP hourly limit to be exceeded")
	}
	if result.Reason != "IP hourly limit exceeded" {
		t.Errorf("unexpected reason: %s", result.Reason)
	}
}

func TestCheckSMSRequest_InvalidPhoneFormat(t *testing.T) {
	p := NewSMSProtector()

	tests := []string{
		"14155552671",  // missing +
		"+0123456789",  // starts with 0
		"+1",           // too short
		"not-a-number", // garbage
		"",             // empty
	}

	for _, phone := range tests {
		result := p.CheckSMSRequest(phone, "1.2.3.4", "fp-abc")
		if result.Allowed {
			t.Errorf("expected invalid phone %q to be rejected", phone)
		}
	}
}

func TestCheckSMSRequest_DisposableNumber(t *testing.T) {
	p := NewSMSProtector()
	result := p.CheckSMSRequest("+19005551234", "1.2.3.4", "fp-abc")
	if result.Allowed {
		t.Error("expected disposable number to be rejected")
	}
	if result.Reason != "disposable or virtual phone number not allowed" {
		t.Errorf("unexpected reason: %s", result.Reason)
	}
}

func TestRecordSMSCost(t *testing.T) {
	p := NewSMSProtector()

	cost := p.GetDailyCostCents("app-1")
	if cost != 0 {
		t.Errorf("expected 0 cost before any SMS, got %d", cost)
	}

	p.RecordSMSCost("app-1")
	p.RecordSMSCost("app-1")
	p.RecordSMSCost("app-1")

	cost = p.GetDailyCostCents("app-1")
	if cost != 3 {
		t.Errorf("expected 3 cents after 3 SMS, got %d", cost)
	}
}
