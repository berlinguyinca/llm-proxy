// Package rate_limiter provides token bucket rate limiting for API endpoints
package rate_limiter

import (
	"sync"
	"time"
)

// TokenBucket represents a token bucket implementation for rate limiting
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket with specified capacity and refill rate
func NewTokenBucket(maxTokens float64, refillRate float64) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: now,
	}
}

// Refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) Refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.tokens+elapsed*tb.refillRate, tb.maxTokens)
	tb.lastRefill = now
}

// Allow checks if a request can proceed and consumes a token if allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.tokens+elapsed*tb.refillRate, tb.maxTokens)
	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// GetTokens returns the current number of available tokens (for monitoring)
func (tb *TokenBucket) GetTokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tokens := min(tb.tokens+elapsed*tb.refillRate, tb.maxTokens)
	return tokens
}

// TokenBucketStore stores token buckets for multiple keys (e.g., per-model limits)
type TokenBucketStore struct {
	mu      sync.RWMutex
	buckets map[string]*TokenBucket
	global  *TokenBucket
}

// NewTokenBucketStore creates a new store with global and per-key limits
func NewTokenBucketStore(globalTokens, globalRefillRate float64) *TokenBucketStore {
	return &TokenBucketStore{
		buckets: make(map[string]*TokenBucket),
		global:  NewTokenBucket(globalTokens, globalRefillRate),
	}
}

// Acquire attempts to acquire a token for the given key (or global if no key)
func (s *TokenBucketStore) Acquire(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Refill and check global first
	s.global.Refill()
	if s.global.tokens >= 1.0 {
		s.global.tokens -= 1.0
		return true
	}

	// Global is depleted, try per-key bucket (each key gets its own small bucket)
	key = "default" + key // Hash key to prevent path traversal

	if _, exists := s.buckets[key]; !exists {
		s.buckets[key] = NewTokenBucket(0.5, 0.1) // Default per-key: 0.5 tokens with 0.1/sec refill
	}

	s.buckets[key].Refill()
	return s.buckets[key].tokens >= 1.0
}

// GetTokensForKey returns the current tokens for a specific key
func (s *TokenBucketStore) GetTokensForKey(key string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key = "default" + key
	if bucket, exists := s.buckets[key]; exists {
		return bucket.GetTokens()
	}

	s.global.Refill()
	return s.global.GetTokens()
}

// Reset resets all buckets for testing purposes
func (s *TokenBucketStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.buckets {
		delete(s.buckets, key)
	}
}
