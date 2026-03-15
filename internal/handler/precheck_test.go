package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
	challengeEngine "github.com/engagelab/captcha/internal/service/challenge"
	policyEngine "github.com/engagelab/captcha/internal/service/policy"
	riskEngine "github.com/engagelab/captcha/internal/service/risk"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupPrecheckRouter() (*gin.Engine, *repository.MemoryStore) {
	store := repository.NewMemoryStore()
	risk := riskEngine.NewEngine()
	policy := policyEngine.NewEngine()
	challenge := challengeEngine.NewEngine("test-secret")

	h := NewPrecheckHandler(store, risk, policy, challenge)

	r := gin.New()
	r.POST("/v1/risk/precheck", h.Handle)
	return r, store
}

func TestPrecheck_NormalRequest_PassOrInvisible(t *testing.T) {
	r, _ := setupPrecheckRouter()

	body := model.PrecheckRequest{
		AppID:       "app-001",
		IP:          "1.2.3.4",
		UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Fingerprint: "normal-user-fp-12345",
		BehaviorData: map[string]interface{}{
			"mouse_entropy":   3.5,
			"time_on_page_ms": 15000.0,
			"key_count":       42.0,
			"scroll_count":    8.0,
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.PrecheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Normal request should get pass or invisible
	if resp.Action != model.RiskActionPass && resp.Action != model.RiskActionInvisible {
		t.Errorf("expected action pass or invisible for normal request, got %s", resp.Action)
	}

	// For pass/invisible, a token should be returned
	if resp.Action == model.RiskActionPass || resp.Action == model.RiskActionInvisible {
		if resp.Token == "" {
			t.Error("expected non-empty token for pass/invisible action")
		}
	}
}

func TestPrecheck_BotLikeRequest_ChallengeOrDeny(t *testing.T) {
	r, _ := setupPrecheckRouter()

	body := model.PrecheckRequest{
		AppID:       "app-001",
		IP:          "198.51.100.50", // known bad IP + datacenter range
		UserAgent:   "python-requests/2.28.0",
		Fingerprint: "",
		BehaviorData: map[string]interface{}{
			"mouse_entropy":   0.0,
			"time_on_page_ms": 50.0,
			"key_count":       0.0,
			"scroll_count":    0.0,
		},
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.PrecheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Bot-like request should get challenge or deny
	if resp.Action != model.RiskActionChallenge && resp.Action != model.RiskActionDeny {
		t.Errorf("expected action challenge or deny for bot-like request, got %s (score: %f)", resp.Action, resp.RiskScore)
	}
}

func TestPrecheck_MissingAppID_Returns400(t *testing.T) {
	r, _ := setupPrecheckRouter()

	// Empty body (missing required app_id)
	body := map[string]interface{}{
		"ip": "1.2.3.4",
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing app_id, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPrecheck_InvalidAppID_Returns400(t *testing.T) {
	r, _ := setupPrecheckRouter()

	body := model.PrecheckRequest{
		AppID: "nonexistent-app",
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid app_id, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPrecheck_InvalidJSON_Returns400(t *testing.T) {
	r, _ := setupPrecheckRouter()

	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", w.Code)
	}
}

func TestPrecheck_ResponseHasChallengeID(t *testing.T) {
	r, _ := setupPrecheckRouter()

	body := model.PrecheckRequest{
		AppID:       "app-001",
		IP:          "1.2.3.4",
		UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Fingerprint: "fp-test",
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk/precheck", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp model.PrecheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.ChallengeID == "" {
		t.Error("expected non-empty challenge_id in response")
	}
}
