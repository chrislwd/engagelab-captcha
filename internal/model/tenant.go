package model

import "time"

// Plan represents a billing/feature tier for a tenant.
type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

// Tenant represents a customer account that can own multiple apps.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	Plan      Plan      `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User type is defined in user.go
