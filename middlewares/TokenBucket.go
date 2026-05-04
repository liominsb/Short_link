package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket 定义令牌桶结构
type TokenBucket struct {
	rate       float64   // 令牌生成速率（每秒生成的令牌数）
	capacity   float64   // 令牌桶容量
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次填充令牌的时间
	mutex      sync.Mutex
}

// NewTokenBucket 创建一个新的令牌桶
func NewTokenBucket(rate float64, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow 尝试获取一个令牌，返回是否允许
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// 计算新增的令牌数
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1 {
		tb.tokens -= 1
		return true
	}

	return false
}

// RateLimiter 定义限流器结构
type RateLimiter struct {
	clients  map[string]*TokenBucket
	mutex    sync.Mutex
	rate     float64
	capacity float64
}

// NewRateLimiter 创建一个新的限流器
func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		clients:  make(map[string]*TokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

// GetTokenBucket 获取或创建客户端的令牌桶
func (rl *RateLimiter) GetTokenBucket(clientID string) *TokenBucket {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	tb, exists := rl.clients[clientID]
	if !exists {
		tb = NewTokenBucket(rl.rate, rl.capacity)
		rl.clients[clientID] = tb
	}

	return tb
}

// Cleanup 定期清理不活跃的客户端（可选）
func (rl *RateLimiter) Cleanup(timeout time.Duration) {
	for {
		time.Sleep(timeout)
		rl.mutex.Lock()
		for clientID, tb := range rl.clients {
			tb.mutex.Lock()
			if time.Since(tb.lastRefill) > timeout {
				delete(rl.clients, clientID)
			}
			tb.mutex.Unlock()
		}
		rl.mutex.Unlock()
	}
}

// RateLimitMiddleware 返回一个 Gin 中间件，用于限流
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		tb := rl.GetTokenBucket(clientIP)

		if tb.Allow() {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests",
			})
			return
		}
	}
}
