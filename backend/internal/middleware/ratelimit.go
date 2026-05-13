package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type loginAttempt struct {
	count     int
	blockedAt time.Time
	lastTry   time.Time
}

var (
	attempts = map[string]*loginAttempt{}
	mu       sync.Mutex
)

const (
	maxAttempts  = 5               // попыток до блокировки
	blockDur     = 15 * time.Minute // время блокировки
	windowDur    = 10 * time.Minute // окно подсчёта попыток
)

// RateLimit блокирует IP после 5 неудачных попыток входа на 15 минут
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		a, ok := attempts[ip]
		if !ok {
			a = &loginAttempt{}
			attempts[ip] = a
		}

		// Снять блокировку если время вышло
		if !a.blockedAt.IsZero() && time.Since(a.blockedAt) > blockDur {
			a.count = 0
			a.blockedAt = time.Time{}
		}

		// Сбросить счётчик если окно истекло
		if time.Since(a.lastTry) > windowDur {
			a.count = 0
		}

		if !a.blockedAt.IsZero() {
			remaining := blockDur - time.Since(a.blockedAt)
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many attempts, try later",
				"retry_after": int(remaining.Seconds()),
			})
			return
		}
		mu.Unlock()

		c.Next()

		// Если ответ 401 — считаем неудачную попытку
		mu.Lock()
		defer mu.Unlock()
		if c.Writer.Status() == http.StatusUnauthorized {
			a.count++
			a.lastTry = time.Now()
			if a.count >= maxAttempts {
				a.blockedAt = time.Now()
			}
		} else if c.Writer.Status() == http.StatusOK {
			// Успешный вход — сбрасываем счётчик
			a.count = 0
			a.blockedAt = time.Time{}
		}
	}
}
