package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

const testHMACSecret = "test-hmac-secret-key"

func newTestService() *Service {
	store := repository.NewMemoryStore()
	return NewService(store, testHMACSecret)
}

func signPayload(payload string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestSiteVerify_ValidToken(t *testing.T) {
	svc := newTestService()

	token := svc.GenerateToken("challenge-valid-001")

	req := model.SiteVerifyRequest{
		Token:  token,
		Secret: "sec_demo_secret_key_xyz",
	}

	resp := svc.SiteVerify(req)
	if !resp.Success {
		t.Errorf("expected success=true, got false. errors: %v", resp.ErrorCodes)
	}
	if resp.ChallengeTS == "" {
		t.Error("expected non-empty challenge_ts")
	}
}

func TestSiteVerify_InvalidTokenFormat(t *testing.T) {
	svc := newTestService()

	req := model.SiteVerifyRequest{
		Token:  "this-is-not-a-valid-token",
		Secret: "sec_demo_secret_key_xyz",
	}

	resp := svc.SiteVerify(req)
	if resp.Success {
		t.Error("expected success=false for invalid token format")
	}
	if !containsError(resp.ErrorCodes, "invalid-token-format") {
		t.Errorf("expected error 'invalid-token-format', got %v", resp.ErrorCodes)
	}
}

func TestSiteVerify_InvalidSignature(t *testing.T) {
	svc := newTestService()

	token := "fake-challenge|2025-01-01T00:00:00Z.badsignature1234567890abcdef"

	req := model.SiteVerifyRequest{
		Token:  token,
		Secret: "sec_demo_secret_key_xyz",
	}

	resp := svc.SiteVerify(req)
	if resp.Success {
		t.Error("expected success=false for invalid signature")
	}
	if !containsError(resp.ErrorCodes, "invalid-token-signature") {
		t.Errorf("expected error 'invalid-token-signature', got %v", resp.ErrorCodes)
	}
}

func TestSiteVerify_ExpiredToken(t *testing.T) {
	store := repository.NewMemoryStore()
	svc := &Service{
		store:      store,
		hmacKey:    []byte(testHMACSecret),
		usedTokens: make(map[string]time.Time),
	}

	// Create a token with a timestamp > 10 minutes ago
	oldTime := time.Now().UTC().Add(-15 * time.Minute).Format(time.RFC3339)
	payload := fmt.Sprintf("old-challenge|%s", oldTime)
	sig := signPayload(payload, []byte(testHMACSecret))
	token := fmt.Sprintf("%s.%s", payload, sig)

	req := model.SiteVerifyRequest{
		Token:  token,
		Secret: "sec_demo_secret_key_xyz",
	}

	resp := svc.SiteVerify(req)
	if resp.Success {
		t.Error("expected success=false for expired token")
	}
	if !containsError(resp.ErrorCodes, "token-expired") {
		t.Errorf("expected error 'token-expired', got %v", resp.ErrorCodes)
	}
}

func TestSiteVerify_ReplayPrevention(t *testing.T) {
	svc := newTestService()

	token := svc.GenerateToken("challenge-replay-001")

	req := model.SiteVerifyRequest{
		Token:  token,
		Secret: "sec_demo_secret_key_xyz",
	}

	resp1 := svc.SiteVerify(req)
	if !resp1.Success {
		t.Errorf("expected first verification to succeed, errors: %v", resp1.ErrorCodes)
	}

	resp2 := svc.SiteVerify(req)
	if resp2.Success {
		t.Error("expected second verification to fail (replay prevention)")
	}
	if !containsError(resp2.ErrorCodes, "token-already-used") {
		t.Errorf("expected error 'token-already-used', got %v", resp2.ErrorCodes)
	}
}

func TestSiteVerify_InvalidSecret(t *testing.T) {
	svc := newTestService()
	token := svc.GenerateToken("challenge-001")

	req := model.SiteVerifyRequest{
		Token:  token,
		Secret: "invalid-secret-key",
	}

	resp := svc.SiteVerify(req)
	if resp.Success {
		t.Error("expected success=false for invalid secret")
	}
	if !containsError(resp.ErrorCodes, "invalid-secret") {
		t.Errorf("expected error 'invalid-secret', got %v", resp.ErrorCodes)
	}
}

func TestGenerateToken_Format(t *testing.T) {
	svc := newTestService()
	token := svc.GenerateToken("my-challenge-id")

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Token format: "challengeID|timestamp.signature"
	dotIdx := -1
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx == -1 {
		t.Fatalf("expected token to contain '.', got %q", token)
	}

	payload := token[:dotIdx]
	sig := token[dotIdx+1:]

	if payload == "" {
		t.Error("expected non-empty payload")
	}
	if sig == "" {
		t.Error("expected non-empty signature")
	}

	pipeIdx := -1
	for i := 0; i < len(payload); i++ {
		if payload[i] == '|' {
			pipeIdx = i
			break
		}
	}
	if pipeIdx == -1 {
		t.Errorf("expected payload to contain '|', got %q", payload)
	}

	challengeID := payload[:pipeIdx]
	if challengeID != "my-challenge-id" {
		t.Errorf("expected challenge ID 'my-challenge-id', got %q", challengeID)
	}
}

func containsError(codes []string, target string) bool {
	for _, c := range codes {
		if c == target {
			return true
		}
	}
	return false
}
