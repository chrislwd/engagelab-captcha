package model

import "time"

// ChallengeType defines the kind of CAPTCHA challenge presented to the user.
type ChallengeType string

const (
	ChallengeTypeInvisible ChallengeType = "invisible"
	ChallengeTypeSlider    ChallengeType = "slider"
	ChallengeTypeClick     ChallengeType = "click"
	ChallengeTypePuzzle    ChallengeType = "puzzle"
)

// ChallengeStatus tracks the lifecycle of a challenge session.
type ChallengeStatus string

const (
	ChallengeStatusPending ChallengeStatus = "pending"
	ChallengeStatusPassed  ChallengeStatus = "passed"
	ChallengeStatusFailed  ChallengeStatus = "failed"
	ChallengeStatusExpired ChallengeStatus = "expired"
)

// RiskAction is the action the system recommends based on risk assessment.
type RiskAction string

const (
	RiskActionPass      RiskAction = "pass"
	RiskActionInvisible RiskAction = "invisible"
	RiskActionChallenge RiskAction = "challenge"
	RiskActionDeny      RiskAction = "deny"
)

// ChallengeSession represents a single CAPTCHA interaction from request to resolution.
type ChallengeSession struct {
	ID            string          `json:"id"`
	AppID         string          `json:"app_id"`
	SceneID       string          `json:"scene_id"`
	SessionID     string          `json:"session_id"`
	IP            string          `json:"ip"`
	UAHash        string          `json:"ua_hash"`
	FingerprintID string          `json:"fingerprint_id"`
	ChallengeType ChallengeType   `json:"challenge_type"`
	RiskScore     float64         `json:"risk_score"`
	RiskLabel     string          `json:"risk_label"`
	Status        ChallengeStatus `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	ExpiresAt     time.Time       `json:"expires_at"`
}

// ChallengeConfig holds parameters sent to the client to render a challenge.
type ChallengeConfig struct {
	ChallengeID   string        `json:"challenge_id"`
	ChallengeType ChallengeType `json:"challenge_type"`
	Token         string        `json:"token,omitempty"`

	// Slider challenge fields
	SliderBgWidth  int `json:"slider_bg_width,omitempty"`
	SliderBgHeight int `json:"slider_bg_height,omitempty"`
	SliderX        int `json:"slider_x,omitempty"`
	SliderY        int `json:"slider_y,omitempty"`
	SliderWidth    int `json:"slider_width,omitempty"`
	SliderHeight   int `json:"slider_height,omitempty"`

	// Click challenge fields
	ClickTargets []ClickTarget `json:"click_targets,omitempty"`
	ClickPrompt  string        `json:"click_prompt,omitempty"`

	// Puzzle challenge fields
	PuzzleImageURL string `json:"puzzle_image_url,omitempty"`
	PuzzlePieceX   int    `json:"puzzle_piece_x,omitempty"`
	PuzzlePieceY   int    `json:"puzzle_piece_y,omitempty"`

	ExpiresAt time.Time `json:"expires_at"`
}

// ClickTarget represents a single clickable target in a click challenge.
type ClickTarget struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

// RenderChallengeRequest is sent by the SDK to fetch a challenge to display.
type RenderChallengeRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
}

// SubmitChallengeRequest is sent by the SDK when the user completes a challenge.
type SubmitChallengeRequest struct {
	ChallengeID string      `json:"challenge_id" binding:"required"`
	Answer      interface{} `json:"answer" binding:"required"`
}

// SubmitChallengeResponse indicates whether the challenge was passed.
type SubmitChallengeResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}
