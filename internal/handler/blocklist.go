package handler

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BlocklistEntry represents an IP or UA pattern in the blocklist/allowlist.
type BlocklistEntry struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Type      string    `json:"type"`    // "ip" or "ua"
	ListType  string    `json:"list_type"` // "block" or "allow"
	Value     string    `json:"value"`   // IP, CIDR, or UA pattern
	Reason    string    `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type BlocklistStore struct {
	mu      sync.RWMutex
	entries map[string]*BlocklistEntry // id -> entry
}

func NewBlocklistStore() *BlocklistStore {
	return &BlocklistStore{entries: make(map[string]*BlocklistEntry)}
}

// CheckIP returns true if the IP should be blocked.
func (s *BlocklistStore) CheckIP(appID, ip string) (blocked bool, allowed bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parsedIP := net.ParseIP(ip)

	for _, e := range s.entries {
		if e.AppID != appID && e.AppID != "*" {
			continue
		}
		if e.Type != "ip" {
			continue
		}
		if e.ExpiresAt != nil && time.Now().After(*e.ExpiresAt) {
			continue
		}

		match := false
		if e.Value == ip {
			match = true
		} else if _, cidr, err := net.ParseCIDR(e.Value); err == nil && parsedIP != nil {
			match = cidr.Contains(parsedIP)
		}

		if match {
			if e.ListType == "block" {
				return true, false
			}
			if e.ListType == "allow" {
				return false, true
			}
		}
	}
	return false, false
}

// CheckUA returns true if the UA should be blocked.
func (s *BlocklistStore) CheckUA(appID, ua string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.entries {
		if e.AppID != appID && e.AppID != "*" {
			continue
		}
		if e.Type != "ua" || e.ListType != "block" {
			continue
		}
		if e.ExpiresAt != nil && time.Now().After(*e.ExpiresAt) {
			continue
		}
		if containsStr(ua, e.Value) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

type BlocklistHandler struct {
	store *BlocklistStore
}

func NewBlocklistHandler(store *BlocklistStore) *BlocklistHandler {
	return &BlocklistHandler{store: store}
}

// Create handles POST /v1/blocklist
func (h *BlocklistHandler) Create(c *gin.Context) {
	var req struct {
		AppID     string `json:"app_id" binding:"required"`
		Type      string `json:"type" binding:"required"`      // ip, ua
		ListType  string `json:"list_type" binding:"required"` // block, allow
		Value     string `json:"value" binding:"required"`
		Reason    string `json:"reason"`
		TTLHours  int    `json:"ttl_hours,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := &BlocklistEntry{
		ID:        uuid.New().String(),
		AppID:     req.AppID,
		Type:      req.Type,
		ListType:  req.ListType,
		Value:     req.Value,
		Reason:    req.Reason,
		CreatedAt: time.Now(),
	}
	if req.TTLHours > 0 {
		expires := time.Now().Add(time.Duration(req.TTLHours) * time.Hour)
		entry.ExpiresAt = &expires
	}

	h.store.mu.Lock()
	h.store.entries[entry.ID] = entry
	h.store.mu.Unlock()

	c.JSON(http.StatusCreated, entry)
}

// List handles GET /v1/blocklist
func (h *BlocklistHandler) List(c *gin.Context) {
	appID := c.Query("app_id")
	listType := c.Query("list_type")

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var result []*BlocklistEntry
	for _, e := range h.store.entries {
		if appID != "" && e.AppID != appID {
			continue
		}
		if listType != "" && e.ListType != listType {
			continue
		}
		result = append(result, e)
	}
	if result == nil {
		result = []*BlocklistEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"entries": result, "total": len(result)})
}

// Delete handles DELETE /v1/blocklist/:id
func (h *BlocklistHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	if _, ok := h.store.entries[id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	delete(h.store.entries, id)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
