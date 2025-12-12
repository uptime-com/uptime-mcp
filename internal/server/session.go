package server

import (
	"sync"
	"time"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// sessionCache caches validated sessions by token.
type sessionCache struct {
	mu      sync.RWMutex
	entries map[string]*sessionCacheEntry
	ttl     time.Duration
}

type sessionCacheEntry struct {
	session   *app.Session
	expiresAt time.Time
}

func newSessionCache(ttl time.Duration) *sessionCache {
	return &sessionCache{
		entries: make(map[string]*sessionCacheEntry),
		ttl:     ttl,
	}
}

func (c *sessionCache) get(token string) *app.Session {
	c.mu.RLock()
	entry, ok := c.entries[token]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.session
}

func (c *sessionCache) set(token string, session *app.Session) {
	c.mu.Lock()
	c.entries[token] = &sessionCacheEntry{
		session:   session,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
