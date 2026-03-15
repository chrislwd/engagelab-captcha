package repository

import "github.com/engagelab/captcha/internal/model"

// AppStore defines persistence operations for App entities.
type AppStore interface {
	CreateApp(app *model.App) error
	GetApp(id string) (*model.App, error)
	GetAppBySiteKey(siteKey string) (*model.App, error)
	GetAppBySecretKey(secretKey string) (*model.App, error)
	ListAppsByTenant(tenantID string) []*model.App
	DeleteApp(id string) error
}

// SceneStore defines persistence operations for Scene entities.
type SceneStore interface {
	CreateScene(scene *model.Scene) error
	GetScene(id string) (*model.Scene, error)
	ListScenesByApp(appID string) []*model.Scene
}

// ChallengeStore defines persistence operations for ChallengeSession entities.
type ChallengeStore interface {
	SaveChallenge(ch *model.ChallengeSession) error
	GetChallenge(id string) (*model.ChallengeSession, error)
	UpdateChallengeStatus(id string, status model.ChallengeStatus) error
	ListChallenges() []*model.ChallengeSession
}

// PolicyStore defines persistence operations for Policy entities.
type PolicyStore interface {
	CreatePolicy(p *model.Policy) error
	GetPolicy(id string) (*model.Policy, error)
	GetPolicyByScene(sceneType model.SceneType) (*model.Policy, error)
	ListPolicies() []*model.Policy
}

// TenantStore defines persistence operations for Tenant entities.
type TenantStore interface {
	CreateTenant(t *model.Tenant)
	GetTenantByAPIKey(apiKey string) (*model.Tenant, error)
}

// Store combines all entity stores into a single interface.
type Store interface {
	AppStore
	SceneStore
	ChallengeStore
	PolicyStore
	TenantStore
}
