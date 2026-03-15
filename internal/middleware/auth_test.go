package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAPIKeyRouter() (*gin.Engine, *repository.MemoryStore) {
	store := repository.NewMemoryStore()

	r := gin.New()
	r.GET("/protected", APIKeyAuth(store), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"tenant_id":   c.GetString("tenant_id"),
			"tenant_name": c.GetString("tenant_name"),
		})
	})
	return r, store
}

func setupSiteKeyRouter() (*gin.Engine, *repository.MemoryStore) {
	store := repository.NewMemoryStore()

	r := gin.New()
	r.GET("/sdk", SiteKeyAuth(store), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app_id":   c.GetString("app_id"),
			"site_key": c.GetString("site_key"),
		})
	})
	return r, store
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	r, _ := setupAPIKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", "ak_demo_key_123456") // seeded key
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["tenant_id"] != "tenant-001" {
		t.Errorf("expected tenant_id 'tenant-001', got %q", resp["tenant_id"])
	}
	if resp["tenant_name"] != "Demo Corp" {
		t.Errorf("expected tenant_name 'Demo Corp', got %q", resp["tenant_name"])
	}
}

func TestAPIKeyAuth_MissingKey_Returns401(t *testing.T) {
	r, _ := setupAPIKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No X-API-Key header
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for missing API key, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestAPIKeyAuth_InvalidKey_Returns401(t *testing.T) {
	r, _ := setupAPIKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", "invalid-key-that-does-not-exist")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid API key, got %d", w.Code)
	}
}

func TestSiteKeyAuth_ValidSiteKey_Header(t *testing.T) {
	r, _ := setupSiteKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/sdk", nil)
	req.Header.Set("X-Site-Key", "sk_demo_site_key_abc") // seeded site key
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["app_id"] != "app-001" {
		t.Errorf("expected app_id 'app-001', got %q", resp["app_id"])
	}
}

func TestSiteKeyAuth_ValidSiteKey_BearerAuth(t *testing.T) {
	r, _ := setupSiteKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/sdk", nil)
	req.Header.Set("Authorization", "Bearer sk_demo_site_key_abc")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["app_id"] != "app-001" {
		t.Errorf("expected app_id 'app-001', got %q", resp["app_id"])
	}
}

func TestSiteKeyAuth_InvalidSiteKey_Returns401(t *testing.T) {
	r, _ := setupSiteKeyRouter()

	req := httptest.NewRequest(http.MethodGet, "/sdk", nil)
	req.Header.Set("X-Site-Key", "invalid-site-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid site key, got %d", w.Code)
	}
}

func TestSiteKeyAuth_NoSiteKey_PassesThrough(t *testing.T) {
	r, _ := setupSiteKeyRouter()

	// SiteKeyAuth allows requests through when no site key is provided
	req := httptest.NewRequest(http.MethodGet, "/sdk", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for missing site key (pass-through), got %d", w.Code)
	}

	// app_id should be empty since no site key was provided
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["app_id"] != "" {
		t.Errorf("expected empty app_id when no site key provided, got %q", resp["app_id"])
	}
}
