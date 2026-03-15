package model

import "time"

// SceneType categorizes the context in which a CAPTCHA challenge is presented.
type SceneType string

const (
	SceneTypeRegister SceneType = "register"
	SceneTypeLogin    SceneType = "login"
	SceneTypeActivity SceneType = "activity"
	SceneTypeComment  SceneType = "comment"
	SceneTypeAPI      SceneType = "api"
)

// SceneStatus indicates whether a scene is enabled.
type SceneStatus string

const (
	SceneStatusActive   SceneStatus = "active"
	SceneStatusInactive SceneStatus = "inactive"
)

// Scene represents a specific use-case context for CAPTCHA within an app.
type Scene struct {
	ID        string      `json:"id"`
	AppID     string      `json:"app_id"`
	SceneType SceneType   `json:"scene_type"`
	PolicyID  string      `json:"policy_id"`
	Status    SceneStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CreateSceneRequest is the payload for creating a new scene.
type CreateSceneRequest struct {
	SceneType SceneType `json:"scene_type" binding:"required"`
	PolicyID  string    `json:"policy_id"`
}

// CreateSceneResponse is returned after creating a scene.
type CreateSceneResponse struct {
	ID        string      `json:"id"`
	AppID     string      `json:"app_id"`
	SceneType SceneType   `json:"scene_type"`
	PolicyID  string      `json:"policy_id"`
	Status    SceneStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}
