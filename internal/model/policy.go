package model

import "time"

// Policy defines the risk thresholds and actions for a given scene type.
type Policy struct {
	ID            string     `json:"id"`
	SceneType     SceneType  `json:"scene_type"`
	ThresholdLow  float64    `json:"threshold_low"`
	ThresholdHigh float64    `json:"threshold_high"`
	ActionLow     RiskAction `json:"action_low"`
	ActionMid     RiskAction `json:"action_mid"`
	ActionHigh    RiskAction `json:"action_high"`
	IPWhitelist   []string   `json:"ip_whitelist"`
	IPBlacklist   []string   `json:"ip_blacklist"`
	RateLimitRPM  int        `json:"rate_limit_rpm"`
	RateLimitRPH  int        `json:"rate_limit_rph"`
	Enabled       bool       `json:"enabled"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreatePolicyRequest is the payload for creating or updating a policy.
type CreatePolicyRequest struct {
	SceneType     SceneType  `json:"scene_type" binding:"required"`
	ThresholdLow  float64    `json:"threshold_low"`
	ThresholdHigh float64    `json:"threshold_high"`
	ActionLow     RiskAction `json:"action_low"`
	ActionMid     RiskAction `json:"action_mid"`
	ActionHigh    RiskAction `json:"action_high"`
	IPWhitelist   []string   `json:"ip_whitelist"`
	IPBlacklist   []string   `json:"ip_blacklist"`
	RateLimitRPM  int        `json:"rate_limit_rpm"`
	RateLimitRPH  int        `json:"rate_limit_rph"`
}
