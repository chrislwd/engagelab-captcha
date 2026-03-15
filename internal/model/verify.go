package model

import "time"

// VerifyResult records the outcome of a challenge verification.
type VerifyResult struct {
	ID          string    `json:"id"`
	ChallengeID string    `json:"challenge_id"`
	Verified    bool      `json:"verified"`
	Score       float64   `json:"score"`
	Labels      []string  `json:"labels"`
	Action      string    `json:"action"`
	ReasonCode  string    `json:"reason_code"`
	CompletedAt time.Time `json:"completed_at"`
}

// SiteVerifyRequest is sent by the customer's backend to validate a CAPTCHA token.
type SiteVerifyRequest struct {
	Token  string `json:"token" binding:"required"`
	Secret string `json:"secret" binding:"required"`
}

// SiteVerifyResponse is returned to the customer's backend after token validation.
type SiteVerifyResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	Labels      []string `json:"labels"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// PrecheckRequest is sent by the SDK before displaying a challenge.
type PrecheckRequest struct {
	AppID        string                 `json:"app_id" binding:"required"`
	SceneID      string                 `json:"scene_id"`
	IP           string                 `json:"ip"`
	UserAgent    string                 `json:"ua"`
	Fingerprint  string                 `json:"fingerprint"`
	BehaviorData map[string]interface{} `json:"behavior_data"`
}

// PrecheckResponse tells the SDK what action to take.
type PrecheckResponse struct {
	Action        RiskAction    `json:"action"`
	RiskScore     float64       `json:"risk_score"`
	ChallengeType ChallengeType `json:"challenge_type,omitempty"`
	ChallengeID   string        `json:"challenge_id,omitempty"`
	Token         string        `json:"token,omitempty"`
}

// EventFeedbackRequest allows customers to report false positives/negatives.
type EventFeedbackRequest struct {
	ChallengeID  string `json:"challenge_id" binding:"required"`
	FeedbackType string `json:"feedback_type" binding:"required"`
	Comment      string `json:"comment"`
}

// EventFeedback stores submitted feedback.
type EventFeedback struct {
	ID           string    `json:"id"`
	ChallengeID  string    `json:"challenge_id"`
	FeedbackType string    `json:"feedback_type"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
}

// DashboardStats provides aggregate statistics for the console.
type DashboardStats struct {
	TotalChallenges  int64              `json:"total_challenges"`
	ChallengeRate    float64            `json:"challenge_rate"`
	PassRate         float64            `json:"pass_rate"`
	DenyRate         float64            `json:"deny_rate"`
	RiskDistribution map[string]int64   `json:"risk_distribution"`
	TopCountries     []CountryStat      `json:"top_countries"`
}

// CountryStat holds challenge counts per country.
type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}
