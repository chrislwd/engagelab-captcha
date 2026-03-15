package model

import "time"

// AppStatus indicates whether an app is currently active.
type AppStatus string

const (
	AppStatusActive   AppStatus = "active"
	AppStatusInactive AppStatus = "inactive"
)

// App represents a CAPTCHA-enabled application belonging to a tenant.
type App struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Name           string    `json:"name"`
	SiteKey        string    `json:"site_key"`
	SecretKey      string    `json:"secret_key"`
	AllowedDomains []string  `json:"allowed_domains"`
	Status         AppStatus `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateAppRequest is the payload for creating a new app.
type CreateAppRequest struct {
	Name           string   `json:"name" binding:"required"`
	AllowedDomains []string `json:"allowed_domains"`
}

// CreateAppResponse is returned after creating an app.
type CreateAppResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SiteKey   string    `json:"site_key"`
	SecretKey string    `json:"secret_key"`
	Status    AppStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
