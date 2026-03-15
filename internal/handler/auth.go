package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/engagelab/captcha/internal/model"
	"github.com/engagelab/captcha/internal/repository"
)

type AuthHandler struct {
	store *repository.MemoryStore
	mu    sync.RWMutex
	users map[string]*model.User // email -> user
	roles map[string]model.Role  // userID:tenantID -> role
	keys  []apiKeyRecord
}

type apiKeyRecord struct {
	ID        string
	TenantID  string
	Name      string
	KeyPrefix string
	KeyHash   string
	CreatedAt time.Time
}

func NewAuthHandler(store *repository.MemoryStore) *AuthHandler {
	return &AuthHandler{
		store: store,
		users: make(map[string]*model.User),
		roles: make(map[string]model.Role),
	}
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	if _, exists := h.users[req.Email]; exists {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	h.mu.Unlock()

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	// Create tenant
	tenantID := uuid.New().String()
	apiKey := generateKey("ak_")
	siteKey := generateKey("sk_")
	secretKey := generateKey("sec_")

	tenant := &model.Tenant{
		ID:     tenantID,
		Name:   req.Company,
		APIKey: apiKey,
		Plan:   model.PlanFree,
	}

	h.store.CreateTenant(tenant)

	// Create default app
	app := &model.App{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		Name:           req.Company + " App",
		SiteKey:        siteKey,
		SecretKey:      secretKey,
		AllowedDomains: []string{"localhost", "*"},
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	h.store.CreateApp(app)

	h.mu.Lock()
	h.users[req.Email] = user
	h.roles[user.ID+":"+tenantID] = model.RoleOwner
	h.mu.Unlock()

	// Generate JWT-like token (simplified)
	token := generateKey("tok_")

	c.JSON(http.StatusCreated, model.AuthResponse{
		Token:    token,
		User:     user,
		TenantID: tenantID,
		Role:     model.RoleOwner,
	})
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.RLock()
	user, exists := h.users[req.Email]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Find tenant
	var tenantID string
	var role model.Role
	h.mu.RLock()
	for key, r := range h.roles {
		uid := key[:len(user.ID)]
		if uid == user.ID {
			tenantID = key[len(user.ID)+1:]
			role = r
			break
		}
	}
	h.mu.RUnlock()

	token := generateKey("tok_")

	c.JSON(http.StatusOK, model.AuthResponse{
		Token:    token,
		User:     user,
		TenantID: tenantID,
		Role:     role,
	})
}

// GenerateAPIKey handles POST /v1/account/api-keys
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := generateKey("ak_")
	keyHash := hashKey(key)
	record := apiKeyRecord{
		ID:        uuid.New().String(),
		Name:      req.Name,
		KeyPrefix: key[:12],
		KeyHash:   keyHash,
		CreatedAt: time.Now(),
	}

	h.mu.Lock()
	h.keys = append(h.keys, record)
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"id":         record.ID,
		"name":       record.Name,
		"key":        key, // Only shown once
		"key_prefix": record.KeyPrefix,
	})
}

// ListAPIKeys handles GET /v1/account/api-keys
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []model.APIKeyInfo
	for _, k := range h.keys {
		result = append(result, model.APIKeyInfo{
			ID: k.ID, Name: k.Name,
			KeyPrefix: k.KeyPrefix, CreatedAt: k.CreatedAt,
		})
	}
	if result == nil {
		result = []model.APIKeyInfo{}
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": result})
}

func generateKey(prefix string) string {
	b := make([]byte, 24)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)
}
