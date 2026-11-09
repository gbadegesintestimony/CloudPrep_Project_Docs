package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// IPRateLimiter provides thread-safe per-IP token-bucket rate limiting with automatic
// eviction of stale IP entries to prevent memory leaks during flood attacks.
type IPRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*tokenBucket
	rps         float64
	burst       float64
	lastCleanup time.Time
}

func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	if rps <= 0 {
		rps = 20.0
	}
	if burst <= 0 {
		burst = 40
	}
	return &IPRateLimiter{
		buckets:     make(map[string]*tokenBucket),
		rps:         rps,
		burst:       float64(burst),
		lastCleanup: time.Now(),
	}
}

func (l *IPRateLimiter) Allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	// Evict stale IP entries every 3 minutes (or if map grows beyond 10,000 entries)
	if now.Sub(l.lastCleanup) > 3*time.Minute || len(l.buckets) > 10000 {
		cutoff := now.Add(-5 * time.Minute)
		for k, b := range l.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.lastCleanup = now
	}

	b, exists := l.buckets[ip]
	if !exists {
		l.buckets[ip] = &tokenBucket{
			tokens:     l.burst - 1,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * l.rps
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		b.lastRefill = now
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

// Middleware returns a Gin handler that enforces the per-IP token bucket limit.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		if !l.Allow(ip) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "TOO_MANY_REQUESTS",
				"message": "Rate limit exceeded. Please slow down and try again shortly.",
			})
			return
		}
		c.Next()
	}
}

// MaxBodyBytesMiddleware caps incoming request bodies to maxBytes (e.g., 64 KB)
// to prevent memory exhaustion / oversized JSON payload attacks.
func MaxBodyBytesMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "PAYLOAD_TOO_LARGE",
				"message": "Request body exceeds maximum allowed size.",
			})
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware attaches defensive HTTP response headers to every API response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "0")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

// StrictCORSMiddleware enforces an explicit origin allowlist, always sets Vary: Origin,
// never combines wildcard "*" with Access-Control-Allow-Credentials: true, and rejects
// preflight OPTIONS requests from untrusted origins.
func StrictCORSMiddleware(frontendOrigins []string) gin.HandlerFunc {
	allowed := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
	}
	for _, o := range frontendOrigins {
		if o != "" && o != "*" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		c.Header("Vary", "Origin")
		origin := c.GetHeader("Origin")

		if origin != "" {
			if allowed[origin] {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Max-Age", "43200")
			} else if c.Request.Method == http.MethodOptions {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "FORBIDDEN_ORIGIN",
					"message": "Origin is not permitted by CORS policy.",
				})
				return
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

