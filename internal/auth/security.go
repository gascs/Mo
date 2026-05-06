package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders sets standard security headers on all responses.
func SecurityHeaders(domain string) gin.HandlerFunc {
	csp := "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'"

	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", csp)
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "0")

		if domain != "" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// --- Global API rate limiter: 1000 requests per minute per IP ---

type rateBucket struct {
	tokens   float64
	lastSeen time.Time
}

var (
	apiLimiter     = make(map[string]*rateBucket)
	apiLimiterMu   sync.Mutex
	apiRate        = 1000.0       // tokens per minute
	apiBurst       = 1000.0       // max burst
	apiRefillRate  = apiRate / 60.0 // tokens per second
)

func init() {
	go func() {
		for range time.Tick(5 * time.Minute) {
			apiLimiterMu.Lock()
			cutoff := time.Now().Add(-1 * time.Minute)
			for k, v := range apiLimiter {
				if v.lastSeen.Before(cutoff) {
					delete(apiLimiter, k)
				}
			}
			apiLimiterMu.Unlock()
		}
	}()
}

// GlobalRateLimit enforces 1000 req/min per IP on API routes.
func GlobalRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		apiLimiterMu.Lock()
		b, ok := apiLimiter[ip]
		if !ok {
			b = &rateBucket{tokens: apiBurst, lastSeen: now}
			apiLimiter[ip] = b
		}

		// Refill tokens based on elapsed time
		elapsed := now.Sub(b.lastSeen).Seconds()
		b.tokens += elapsed * apiRefillRate
		if b.tokens > apiBurst {
			b.tokens = apiBurst
		}
		b.lastSeen = now

		if b.tokens < 1 {
			apiLimiterMu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "Too many requests"},
			})
			return
		}
		b.tokens--
		apiLimiterMu.Unlock()

		c.Next()
	}
}
