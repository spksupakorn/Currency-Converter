package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/pkg/response"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimit(cfg config.Config) gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)
	cleanupTicker := time.NewTicker(5 * time.Minute)

	go func() {
		for range cleanupTicker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	newVisitor := func() *rate.Limiter {
		per := cfg.RateLimitWindow
		req := cfg.RateLimitRequests
		return rate.NewLimiter(rate.Every(per/time.Duration(req)), req)
	}

	getVisitor := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		v, ok := visitors[ip]
		if !ok {
			lim := newVisitor()
			visitors[ip] = &visitor{limiter: lim, lastSeen: time.Now()}
			return lim
		}
		v.lastSeen = time.Now()
		return v.limiter
	}

	return func(c *gin.Context) {
		lim := getVisitor(c.ClientIP())
		if !lim.Allow() {
			response.TooManyRequests(c, "rate_limited", "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
