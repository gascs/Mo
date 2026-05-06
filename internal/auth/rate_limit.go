package auth

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count    int
	lockedUntil time.Time
}

var (
	rateStore = make(map[string]*rateEntry)
	rateMu    sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rateMu.Lock()
			now := time.Now()
			for k, v := range rateStore {
				// Remove entries with expired lockout
				if v.lockedUntil.Before(now) {
					delete(rateStore, k)
				}
			}
			rateMu.Unlock()
		}
	}()
}

func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateMu.Lock()
		defer rateMu.Unlock()

		entry, exists := rateStore[ip]
		now := time.Now()

		if exists && entry.lockedUntil.After(now) {
			c.AbortWithStatusJSON(429, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "Too many login attempts. Try again in 15 minutes."},
			})
			return
		}

		if !exists {
			rateStore[ip] = &rateEntry{count: 1}
		} else {
			entry.count++
			if entry.count > 5 {
				entry.lockedUntil = now.Add(15 * time.Minute)
				c.AbortWithStatusJSON(429, gin.H{
					"error": gin.H{"code": "RATE_LIMITED", "message": "Too many login attempts. Try again in 15 minutes."},
				})
				return
			}
		}

		c.Next()
	}
}
