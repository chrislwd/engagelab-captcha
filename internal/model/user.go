package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleDev     Role = "developer"
	RoleAnalyst Role = "analyst"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Company  string `json:"company" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	User     *User  `json:"user"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
}

type APIKeyInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}
