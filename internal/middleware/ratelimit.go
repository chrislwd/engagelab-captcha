package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig controls the behaviour of the rate-limiting middleware.
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per window.
	RequestsPerMinute int

	// KeyFunc extracts the rate-limit key from the request. When nil the
	// middleware defaults to the client IP address.
	KeyFunc func(c *gin.Context) string

	// RedisClient is an optional go-redis client. When non-nil the middleware
	// uses a Redis-backed sliding window counter for distributed limiting.
	// When nil (or when Redis is unreachable) it falls back to an in-memory
	// counter (fail-open).
	RedisClient *redis.Client
}

// memEntry holds the state for a single key in the in-memory fallback limiter.
type memEntry struct {
	count    int
	windowAt time.Time
}

// memLimiter is a process-local fallback used when Redis is unavailable.
type memLimiter struct {
	mu      sync.Mutex
	entries map[string]*memEntry
}

func newMemLimiter() *memLimiter {
	return &memLimiter{entries: make(map[string]*memEntry)}
}

// allow returns (allowed, remaining, resetUnix) for the given key.
func (m *memLimiter) allow(key string, limit int, window time.Duration) (bool, int, int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	e, ok := m.entries[key]
	if !ok || now.Sub(e.windowAt) >= window {
		// New window.
		e = &memEntry{count: 0, windowAt: now.Truncate(window)}
		m.entries[key] = e
	}

	resetAt := e.windowAt.Add(window).Unix()
	remaining := limit - e.count

	if e.count >= limit {
		return false, 0, resetAt
	}

	e.count++
	remaining = limit - e.count
	return true, remaining, resetAt
}

// RateLimit returns Gin middleware that enforces per-key request rate limits.
//
// It first attempts to use Redis (sliding window counter). If Redis is not
// configured or temporarily unreachable the middleware transparently falls back
// to a process-local in-memory counter so that the service remains available
// (fail-open).
//
// On every response the following headers are set:
//
//	X-RateLimit-Limit     – the configured limit
//	X-RateLimit-Remaining – requests left in the current window
//	X-RateLimit-Reset     – Unix timestamp when the window resets
//
// When the limit is exceeded the middleware responds with HTTP 429 and a JSON
// body containing retry_after_secs.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60 // sensible default
	}

	window := time.Minute
	fallback := newMemLimiter()

	return func(c *gin.Context) {
		key := c.ClientIP()
		if cfg.KeyFunc != nil {
			key = cfg.KeyFunc(c)
		}

		limit := cfg.RequestsPerMinute
		var allowed bool
		var remaining int
		var resetUnix int64

		if cfg.RedisClient != nil {
			var err error
			allowed, remaining, resetUnix, err = redisAllow(c.Request.Context(), cfg.RedisClient, key, limit, window)
			if err != nil {
				// Redis unavailable – fall back to in-memory.
				allowed, remaining, resetUnix = fallback.allow(key, limit, window)
			}
		} else {
			allowed, remaining, resetUnix = fallback.allow(key, limit, window)
		}

		// Always set rate-limit headers.
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetUnix, 10))

		if !allowed {
			retryAfter := resetUnix - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":           "rate limit exceeded",
				"retry_after_secs": retryAfter,
			})
			return
		}

		c.Next()
	}
}

// IPRateLimit is a convenience constructor that rate-limits by client IP.
func IPRateLimit(rpm int, rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		RequestsPerMinute: rpm,
		RedisClient:       rdb,
	})
}

// FingerprintRateLimit is a convenience constructor that rate-limits by a
// fingerprint ID extracted from the X-Fingerprint-ID header.
func FingerprintRateLimit(rpm int, rdb *redis.Client) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		RequestsPerMinute: rpm,
		RedisClient:       rdb,
		KeyFunc: func(c *gin.Context) string {
			fp := c.GetHeader("X-Fingerprint-ID")
			if fp == "" {
				// Fall back to IP when no fingerprint is provided.
				return "ip:" + c.ClientIP()
			}
			return "fp:" + fp
		},
	})
}

// ---------------------------------------------------------------------------
// Redis sliding window implementation
// ---------------------------------------------------------------------------

// redisAllow uses a sorted-set sliding window to track requests per key.
// Each request is added as a member scored by its Unix-microsecond timestamp.
// Expired members are pruned on every call.
func redisAllow(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (bool, int, int64, error) {
	redisKey := fmt.Sprintf("rl:%s", key)
	now := time.Now()
	windowStart := now.Add(-window)
	resetUnix := now.Add(window).Unix()

	// Use a pipeline for atomicity.
	pipe := rdb.Pipeline()

	// Remove entries outside the window.
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart.UnixMicro(), 10))

	// Count current entries.
	countCmd := pipe.ZCard(ctx, redisKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := int(countCmd.Val())
	remaining := limit - count

	if count >= limit {
		return false, 0, resetUnix, nil
	}

	// Add the new request.
	member := fmt.Sprintf("%d", now.UnixMicro())
	pipe2 := rdb.Pipeline()
	pipe2.ZAdd(ctx, redisKey, redis.Z{Score: float64(now.UnixMicro()), Member: member})
	pipe2.Expire(ctx, redisKey, window+time.Second) // TTL slightly beyond window
	_, err = pipe2.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	remaining = limit - count - 1
	return true, remaining, resetUnix, nil
}
