package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
	"github.com/google/uuid"
)

// Service handles server-side verification of CAPTCHA tokens.
type Service struct {
	mu       sync.Mutex
	store    *repository.MemoryStore
	hmacKey  []byte
	// Replay prevention: set of already-verified tokens.
	usedTokens map[string]time.Time
}

// NewService creates a new verification service.
func NewService(store *repository.MemoryStore, hmacSecret string) *Service {
	s := &Service{
		store:      store,
		hmacKey:    []byte(hmacSecret),
		usedTokens: make(map[string]time.Time),
	}
	// Start background cleanup of expired used-token entries.
	go s.cleanupLoop()
	return s
}

// SiteVerify validates a CAPTCHA token submitted by the customer's backend.
func (s *Service) SiteVerify(req model.SiteVerifyRequest) model.SiteVerifyResponse {
	resp := model.SiteVerifyResponse{
		Success: false,
	}

	// Validate the secret key belongs to a registered app.
	app, err := s.store.GetAppBySecretKey(req.Secret)
	if err != nil {
		resp.ErrorCodes = append(resp.ErrorCodes, "invalid-secret")
		return resp
	}

	// Parse the token: format is "challengeID|timestamp.signature"
	parts := strings.SplitN(req.Token, ".", 2)
	if len(parts) != 2 {
		resp.ErrorCodes = append(resp.ErrorCodes, "invalid-token-format")
		return resp
	}

	payload := parts[0]
	signature := parts[1]

	// Verify HMAC signature.
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		resp.ErrorCodes = append(resp.ErrorCodes, "invalid-token-signature")
		return resp
	}

	// Check for replay.
	s.mu.Lock()
	if _, used := s.usedTokens[req.Token]; used {
		s.mu.Unlock()
		resp.ErrorCodes = append(resp.ErrorCodes, "token-already-used")
		return resp
	}
	s.usedTokens[req.Token] = time.Now()
	s.mu.Unlock()

	// Parse payload components.
	payloadParts := strings.SplitN(payload, "|", 2)
	if len(payloadParts) != 2 {
		resp.ErrorCodes = append(resp.ErrorCodes, "invalid-token-payload")
		return resp
	}

	challengeID := payloadParts[0]
	tokenTS := payloadParts[1]

	// Verify token is not too old (10 minute window).
	ts, err := time.Parse(time.RFC3339, tokenTS)
	if err != nil {
		resp.ErrorCodes = append(resp.ErrorCodes, "invalid-token-timestamp")
		return resp
	}
	if time.Since(ts) > 10*time.Minute {
		resp.ErrorCodes = append(resp.ErrorCodes, "token-expired")
		return resp
	}

	// Look up the challenge session for metadata.
	challenge, err := s.store.GetChallenge(challengeID)
	var score float64
	var action string
	var labels []string
	if err == nil {
		score = challenge.RiskScore
		action = string(challenge.ChallengeType)
		if challenge.RiskLabel != "" {
			labels = strings.Split(challenge.RiskLabel, ",")
		}
	} else {
		// Token is valid even if challenge record is missing (e.g. invisible).
		score = 0
		action = "invisible"
	}

	// Record the verification result.
	vr := &model.VerifyResult{
		ID:          uuid.NewString(),
		ChallengeID: challengeID,
		Verified:    true,
		Score:       score,
		Labels:      labels,
		Action:      action,
		ReasonCode:  "ok",
		CompletedAt: time.Now(),
	}
	_ = s.store.SaveVerifyResult(vr)

	// Determine hostname from allowed domains.
	hostname := ""
	if len(app.AllowedDomains) > 0 {
		hostname = app.AllowedDomains[0]
	}

	resp.Success = true
	resp.Score = score
	resp.Action = action
	resp.Labels = labels
	resp.ChallengeTS = tokenTS
	resp.Hostname = hostname

	return resp
}

// GenerateToken creates a signed token for a challenge (used internally).
func (s *Service) GenerateToken(challengeID string) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	payload := fmt.Sprintf("%s|%s", challengeID, ts)
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", payload, sig)
}

func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		cutoff := time.Now().Add(-15 * time.Minute)
		for token, ts := range s.usedTokens {
			if ts.Before(cutoff) {
				delete(s.usedTokens, token)
			}
		}
		s.mu.Unlock()
	}
}
