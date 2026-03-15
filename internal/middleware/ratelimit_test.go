package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRateLimitRouter(rpm int) *gin.Engine {
	r := gin.New()
	r.Use(RateLimit(RateLimitConfig{
		RequestsPerMinute: rpm,
	}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRateLimit_UnderLimit_AllowsThrough(t *testing.T) {
	r := setupRateLimitRouter(5)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the response body.
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["ok"] != true {
		t.Errorf("expected ok=true in body, got %v", body["ok"])
	}
}

func TestRateLimit_OverLimit_Returns429(t *testing.T) {
	limit := 3
	r := setupRateLimitRouter(limit)

	var lastW *httptest.ResponseRecorder

	// Send limit+1 requests to trigger 429 on the last one.
	for i := 0; i <= limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		lastW = httptest.NewRecorder()
		r.ServeHTTP(lastW, req)
	}

	if lastW.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d: %s", lastW.Code, lastW.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(lastW.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse 429 response body: %v", err)
	}

	if body["error"] != "rate limit exceeded" {
		t.Errorf("expected error='rate limit exceeded', got %q", body["error"])
	}

	retryAfter, ok := body["retry_after_secs"].(float64)
	if !ok || retryAfter <= 0 {
		t.Errorf("expected positive retry_after_secs, got %v", body["retry_after_secs"])
	}
}

func TestRateLimit_Headers_AreSetCorrectly(t *testing.T) {
	limit := 10
	r := setupRateLimitRouter(limit)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.3:12345"
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// X-RateLimit-Limit
	limitHeader := w.Header().Get("X-RateLimit-Limit")
	if limitHeader != strconv.Itoa(limit) {
		t.Errorf("expected X-RateLimit-Limit=%d, got %q", limit, limitHeader)
	}

	// X-RateLimit-Remaining should be limit - 1 after one request.
	remainingHeader := w.Header().Get("X-RateLimit-Remaining")
	expected := strconv.Itoa(limit - 1)
	if remainingHeader != expected {
		t.Errorf("expected X-RateLimit-Remaining=%s, got %q", expected, remainingHeader)
	}

	// X-RateLimit-Reset should be a valid positive unix timestamp.
	resetHeader := w.Header().Get("X-RateLimit-Reset")
	resetVal, err := strconv.ParseInt(resetHeader, 10, 64)
	if err != nil {
		t.Fatalf("X-RateLimit-Reset is not a valid integer: %q", resetHeader)
	}
	if resetVal <= 0 {
		t.Errorf("expected positive X-RateLimit-Reset timestamp, got %d", resetVal)
	}
}

func TestRateLimit_DifferentIPs_IndependentLimits(t *testing.T) {
	limit := 2
	r := setupRateLimitRouter(limit)

	// Exhaust limit for IP-A.
	for i := 0; i < limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.4:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d from IP-A: expected 200, got %d", i, w.Code)
		}
	}

	// IP-B should still be allowed.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.5:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IP-B should not be affected by IP-A limit, got status %d", w.Code)
	}
}

func TestRateLimit_FingerprintKeyFunc(t *testing.T) {
	limit := 2
	r := gin.New()
	r.Use(FingerprintRateLimit(limit, nil))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Two requests with same fingerprint should be allowed.
	for i := 0; i < limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("X-Fingerprint-ID", "fp-abc-123")
		req.RemoteAddr = "10.0.0.6:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	// Third request with same fingerprint should be rejected.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Fingerprint-ID", "fp-abc-123")
	req.RemoteAddr = "10.0.0.6:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for fingerprint over limit, got %d", w.Code)
	}
}
