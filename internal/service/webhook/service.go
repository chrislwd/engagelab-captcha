package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Event types emitted by the CAPTCHA platform.
const (
	EventChallengeCreated  = "challenge.created"
	EventChallengePassed   = "challenge.passed"
	EventChallengeFailed   = "challenge.failed"
	EventRiskHighDetected  = "risk.high_detected"
	EventRiskDenyTriggered = "risk.deny_triggered"
	EventBotDetected       = "bot.detected"
	EventFeedbackReceived  = "feedback.received"
	EventRateLimitHit      = "rate_limit.hit"
)

type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	AppID     string      `json:"app_id"`
	Data      interface{} `json:"data"`
}

type Subscription struct {
	ID       string   `json:"id"`
	AppID    string   `json:"app_id"`
	URL      string   `json:"url"`
	Secret   string   `json:"-"`
	Events   []string `json:"events"`
	Active   bool     `json:"active"`
}

type Store interface {
	GetSubscriptions(ctx context.Context, appID string, eventType string) ([]*Subscription, error)
}

type Service struct {
	store  Store
	client *http.Client
}

func NewService(store Store) *Service {
	return &Service{
		store:  store,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) Emit(ctx context.Context, event *Event) {
	subs, err := s.store.GetSubscriptions(ctx, event.AppID, event.Type)
	if err != nil || len(subs) == 0 {
		return
	}
	for _, sub := range subs {
		go s.deliver(sub, event)
	}
}

func (s *Service) deliver(sub *Subscription, event *Event) {
	body, _ := json.Marshal(event)

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}

		req, err := http.NewRequest(http.MethodPost, sub.URL, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Event", event.Type)
		req.Header.Set("X-Webhook-ID", event.ID)

		if sub.Secret != "" {
			mac := hmac.New(sha256.New, []byte(sub.Secret))
			mac.Write(body)
			req.Header.Set("X-Webhook-Signature", hex.EncodeToString(mac.Sum(nil)))
		}

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}
	}
	log.Printf("webhook delivery failed to %s for event %s", sub.URL, event.Type)
}

// MemoryWebhookStore provides in-memory webhook storage.
type MemoryWebhookStore struct {
	mu   sync.RWMutex
	subs []*Subscription
}

func NewMemoryStore() *MemoryWebhookStore {
	return &MemoryWebhookStore{}
}

func (s *MemoryWebhookStore) Add(sub *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs = append(s.subs, sub)
}

func (s *MemoryWebhookStore) GetSubscriptions(_ context.Context, appID string, eventType string) ([]*Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Subscription
	for _, sub := range s.subs {
		if !sub.Active || (sub.AppID != appID && sub.AppID != "*") {
			continue
		}
		for _, e := range sub.Events {
			if e == eventType || e == "*" {
				result = append(result, sub)
				break
			}
		}
	}
	return result, nil
}

func (s *MemoryWebhookStore) List() []*Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Subscription, len(s.subs))
	copy(result, s.subs)
	return result
}

func (s *MemoryWebhookStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subs {
		if sub.ID == id {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return true
		}
	}
	return false
}

// --- Webhook handler ---

type CreateWebhookRequest struct {
	AppID  string   `json:"app_id" binding:"required"`
	URL    string   `json:"url" binding:"required"`
	Secret string   `json:"secret"`
	Events []string `json:"events" binding:"required"`
}

type WebhookResponse struct {
	ID     string   `json:"id"`
	AppID  string   `json:"app_id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

func FormatResponse(sub *Subscription) *WebhookResponse {
	return &WebhookResponse{
		ID: sub.ID, AppID: sub.AppID,
		URL: sub.URL, Events: sub.Events, Active: sub.Active,
	}
}

func FormatList(subs []*Subscription) []*WebhookResponse {
	var result []*WebhookResponse
	for _, s := range subs {
		result = append(result, FormatResponse(s))
	}
	if result == nil {
		result = []*WebhookResponse{}
	}
	return result
}
