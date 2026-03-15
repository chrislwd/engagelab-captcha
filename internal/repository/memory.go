package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/engagelab/captcha/internal/model"
	"github.com/google/uuid"
)

// MemoryStore provides thread-safe in-memory storage for all entities.
type MemoryStore struct {
	mu sync.RWMutex

	tenants    map[string]*model.Tenant
	users      map[string]*model.User
	apps       map[string]*model.App
	scenes     map[string]*model.Scene
	policies   map[string]*model.Policy
	challenges map[string]*model.ChallengeSession
	results    map[string]*model.VerifyResult
	feedback   map[string]*model.EventFeedback

	// Index: api_key -> tenant_id
	apiKeyIndex map[string]string
	// Index: site_key -> app_id
	siteKeyIndex map[string]string
	// Index: secret_key -> app_id
	secretKeyIndex map[string]string
}

// NewMemoryStore creates a new in-memory store pre-populated with seed data.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		tenants:        make(map[string]*model.Tenant),
		users:          make(map[string]*model.User),
		apps:           make(map[string]*model.App),
		scenes:         make(map[string]*model.Scene),
		policies:       make(map[string]*model.Policy),
		challenges:     make(map[string]*model.ChallengeSession),
		results:        make(map[string]*model.VerifyResult),
		feedback:       make(map[string]*model.EventFeedback),
		apiKeyIndex:    make(map[string]string),
		siteKeyIndex:   make(map[string]string),
		secretKeyIndex: make(map[string]string),
	}
	s.seed()
	return s
}

func (s *MemoryStore) seed() {
	now := time.Now()

	// Seed tenant
	tenant := &model.Tenant{
		ID:        "tenant-001",
		Name:      "Demo Corp",
		APIKey:    "ak_demo_key_123456",
		Plan:      model.PlanPro,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tenants[tenant.ID] = tenant
	s.apiKeyIndex[tenant.APIKey] = tenant.ID

	// Seed user
	user := &model.User{
		ID:           "user-001",
		Email:        "admin@democorp.com",
		PasswordHash: "$2a$10$placeholder",
		Name:         "Admin User",
		Status:       "active",
		CreatedAt:    now,
	}
	s.users[user.ID] = user

	// Seed app
	app := &model.App{
		ID:             "app-001",
		TenantID:       tenant.ID,
		Name:           "Demo Website",
		SiteKey:        "sk_demo_site_key_abc",
		SecretKey:      "sec_demo_secret_key_xyz",
		AllowedDomains: []string{"localhost", "demo.engagelab.cc"},
		Status:         model.AppStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.apps[app.ID] = app
	s.siteKeyIndex[app.SiteKey] = app.ID
	s.secretKeyIndex[app.SecretKey] = app.ID

	// Seed default policies
	policies := []*model.Policy{
		{
			ID:            "policy-login",
			SceneType:     model.SceneTypeLogin,
			ThresholdLow:  20,
			ThresholdHigh: 60,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  30,
			RateLimitRPH:  300,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "policy-register",
			SceneType:     model.SceneTypeRegister,
			ThresholdLow:  15,
			ThresholdHigh: 50,
			ActionLow:     model.RiskActionInvisible,
			ActionMid:     model.RiskActionChallenge,
			ActionHigh:    model.RiskActionDeny,
			RateLimitRPM:  10,
			RateLimitRPH:  60,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "policy-comment",
			SceneType:     model.SceneTypeComment,
			ThresholdLow:  25,
			ThresholdHigh: 70,
			ActionLow:     model.RiskActionPass,
			ActionMid:     model.RiskActionInvisible,
			ActionHigh:    model.RiskActionChallenge,
			RateLimitRPM:  20,
			RateLimitRPH:  200,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
	for _, p := range policies {
		s.policies[p.ID] = p
	}

	// Seed scenes
	scenes := []*model.Scene{
		{
			ID:        "scene-login",
			AppID:     app.ID,
			SceneType: model.SceneTypeLogin,
			PolicyID:  "policy-login",
			Status:    model.SceneStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "scene-register",
			AppID:     app.ID,
			SceneType: model.SceneTypeRegister,
			PolicyID:  "policy-register",
			Status:    model.SceneStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	for _, sc := range scenes {
		s.scenes[sc.ID] = sc
	}

	// Seed some challenge sessions for stats
	statuses := []model.ChallengeStatus{model.ChallengeStatusPassed, model.ChallengeStatusFailed, model.ChallengeStatusPassed, model.ChallengeStatusPassed, model.ChallengeStatusExpired}
	riskScores := []float64{10, 85, 5, 45, 30}
	for i := 0; i < 5; i++ {
		ch := &model.ChallengeSession{
			ID:            fmt.Sprintf("ch-seed-%d", i),
			AppID:         app.ID,
			SceneID:       "scene-login",
			SessionID:     uuid.NewString(),
			IP:            fmt.Sprintf("192.168.1.%d", i+1),
			ChallengeType: model.ChallengeTypeSlider,
			RiskScore:     riskScores[i],
			RiskLabel:     "seed",
			Status:        statuses[i],
			CreatedAt:     now.Add(-time.Duration(i) * time.Hour),
			ExpiresAt:     now.Add(time.Duration(5-i) * time.Minute),
		}
		s.challenges[ch.ID] = ch
	}
}

// --- Tenant ---

// CreateTenant adds a new tenant.
func (s *MemoryStore) CreateTenant(t *model.Tenant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenants[t.ID] = t
	s.apiKeyIndex[t.APIKey] = t.ID
}

// GetTenantByAPIKey looks up a tenant by its API key.
func (s *MemoryStore) GetTenantByAPIKey(apiKey string) (*model.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tid, ok := s.apiKeyIndex[apiKey]
	if !ok {
		return nil, fmt.Errorf("tenant not found for api key")
	}
	t, ok := s.tenants[tid]
	if !ok {
		return nil, fmt.Errorf("tenant not found")
	}
	return t, nil
}

// --- App ---

// CreateApp persists a new app.
func (s *MemoryStore) CreateApp(app *model.App) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apps[app.ID] = app
	s.siteKeyIndex[app.SiteKey] = app.ID
	s.secretKeyIndex[app.SecretKey] = app.ID
	return nil
}

// GetApp returns an app by ID.
func (s *MemoryStore) GetApp(id string) (*model.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("app not found: %s", id)
	}
	return a, nil
}

// GetAppBySiteKey returns an app by its site key.
func (s *MemoryStore) GetAppBySiteKey(siteKey string) (*model.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	aid, ok := s.siteKeyIndex[siteKey]
	if !ok {
		return nil, fmt.Errorf("app not found for site key")
	}
	return s.apps[aid], nil
}

// GetAppBySecretKey returns an app by its secret key.
func (s *MemoryStore) GetAppBySecretKey(secretKey string) (*model.App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	aid, ok := s.secretKeyIndex[secretKey]
	if !ok {
		return nil, fmt.Errorf("app not found for secret key")
	}
	return s.apps[aid], nil
}

// ListAppsByTenant returns all apps for a tenant.
func (s *MemoryStore) ListAppsByTenant(tenantID string) []*model.App {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*model.App
	for _, a := range s.apps {
		if a.TenantID == tenantID {
			result = append(result, a)
		}
	}
	return result
}

// DeleteApp removes an app by ID.
func (s *MemoryStore) DeleteApp(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.apps[id]
	if !ok {
		return fmt.Errorf("app not found: %s", id)
	}
	delete(s.siteKeyIndex, a.SiteKey)
	delete(s.secretKeyIndex, a.SecretKey)
	delete(s.apps, id)
	return nil
}

// --- Scene ---

// CreateScene persists a new scene.
func (s *MemoryStore) CreateScene(scene *model.Scene) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scenes[scene.ID] = scene
	return nil
}

// GetScene returns a scene by ID.
func (s *MemoryStore) GetScene(id string) (*model.Scene, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.scenes[id]
	if !ok {
		return nil, fmt.Errorf("scene not found: %s", id)
	}
	return sc, nil
}

// ListScenesByApp returns all scenes for an app.
func (s *MemoryStore) ListScenesByApp(appID string) []*model.Scene {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*model.Scene
	for _, sc := range s.scenes {
		if sc.AppID == appID {
			result = append(result, sc)
		}
	}
	return result
}

// --- Policy ---

// CreatePolicy persists a new policy.
func (s *MemoryStore) CreatePolicy(p *model.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[p.ID] = p
	return nil
}

// GetPolicy returns a policy by ID.
func (s *MemoryStore) GetPolicy(id string) (*model.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[id]
	if !ok {
		return nil, fmt.Errorf("policy not found: %s", id)
	}
	return p, nil
}

// GetPolicyByScene returns the first enabled policy matching the given scene type.
func (s *MemoryStore) GetPolicyByScene(sceneType model.SceneType) (*model.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.policies {
		if p.SceneType == sceneType && p.Enabled {
			return p, nil
		}
	}
	return nil, fmt.Errorf("policy not found for scene type: %s", sceneType)
}

// ListPolicies returns all policies.
func (s *MemoryStore) ListPolicies() []*model.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*model.Policy
	for _, p := range s.policies {
		result = append(result, p)
	}
	return result
}

// --- Challenge Session ---

// SaveChallenge persists a challenge session.
func (s *MemoryStore) SaveChallenge(ch *model.ChallengeSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.challenges[ch.ID] = ch
	return nil
}

// GetChallenge returns a challenge session by ID.
func (s *MemoryStore) GetChallenge(id string) (*model.ChallengeSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.challenges[id]
	if !ok {
		return nil, fmt.Errorf("challenge not found: %s", id)
	}
	return ch, nil
}

// UpdateChallengeStatus sets the status of an existing challenge.
func (s *MemoryStore) UpdateChallengeStatus(id string, status model.ChallengeStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.challenges[id]
	if !ok {
		return fmt.Errorf("challenge not found: %s", id)
	}
	ch.Status = status
	return nil
}

// ListChallenges returns all stored challenge sessions.
func (s *MemoryStore) ListChallenges() []*model.ChallengeSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*model.ChallengeSession
	for _, ch := range s.challenges {
		result = append(result, ch)
	}
	return result
}

// --- Verify Result ---

// SaveVerifyResult persists a verification result.
func (s *MemoryStore) SaveVerifyResult(vr *model.VerifyResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[vr.ID] = vr
	return nil
}

// --- Feedback ---

// SaveFeedback persists event feedback.
func (s *MemoryStore) SaveFeedback(fb *model.EventFeedback) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedback[fb.ID] = fb
	return nil
}
