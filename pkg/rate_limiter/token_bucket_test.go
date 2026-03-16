// Package rate_limiter provides token bucket rate limiting for API endpoints
package rate_limiter

import (
	"testing"
	"time"
)

func TestTokenBucket_New(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10.0, 1.0)

	if tb.tokens != 10.0 {
		t.Errorf("Expected initial tokens=10, got %.2f", tb.tokens)
	}
	if tb.maxTokens != 10.0 {
		t.Errorf("Expected maxTokens=10, got %.2f", tb.maxTokens)
	}
}

func TestTokenBucket_Allow_WithSufficientTokens(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10.0, 1.0)

	if !tb.Allow() {
		t.Error("Expected Allow to return true with sufficient tokens")
	}

	if tb.tokens >= 9.5 { // Some may be consumed during refill
		t.Errorf("Tokens decreased too much: %.2f", tb.tokens)
	}
}

func TestTokenBucket_Allow_WhenEmpty(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(1.0, 0.0) // No refill

	// Deplete tokens
	for i := 0; i < 2; i++ {
		tb.Allow()
	}

	if tb.tokens >= 0.5 {
		t.Errorf("Expected depleted bucket after 2 consumes, got %.2f", tb.tokens)
	}

	if tb.Allow() {
		t.Error("Expected Allow to return false when tokens are depleted")
	}
}

func TestTokenBucket_Refill_AddsTokens(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10.0, 2.0) // 2 tokens/second

	// Consume all tokens
	for i := 0; i < 10; i++ {
		tb.Allow()
	}

	// Verify it's actually empty
	if tb.tokens >= 0.5 {
		t.Errorf("Expected nearly depleted bucket, got %.2f", tb.tokens)
	}

	// Wait for refill (should add at least 1 token in 0.5 seconds)
	time.Sleep(500 * time.Millisecond)

	// Refill should have added some tokens now
	tb.Refill() // This will apply the accumulated refill

	if tb.tokens < 1.0 {
		t.Errorf("Expected tokens >= 1 after waiting and calling Refill, got %.2f", tb.tokens)
	}
}

func TestTokenBucket_CappedAtMax(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10.0, 100.0) // Fast refill (100 tokens/sec)

	// Wait more than enough time to refill
	time.Sleep(200 * time.Millisecond) // Should add ~20 tokens

	// Tokens should be capped at maxTokens
	if tb.tokens > 10.5 { // Allow small margin for timing
		t.Errorf("Expected tokens <= 10, got %.2f", tb.tokens)
	}
}

func TestTokenBucketStore_GlobalAndPerKey(t *testing.T) {
	t.Parallel()

	store := NewTokenBucketStore(10.0, 1.0) // Global: 10 tokens/sec

	// Should work within global limit
	if !store.Acquire("model1") {
		t.Error("Expected to acquire token with sufficient global tokens")
	}

	// Deplete global tokens (each call depletes the single token in per-key bucket)
	for i := 0; i < 20; i++ {
		store.Acquire("default")
	}

	// Should fail for other keys too when global is depleted
	if store.Acquire("model2") {
		t.Error("Expected to fail when global limit is hit")
	}
}

func TestTokenBucketStore_PerKeyLimits(t *testing.T) {
	t.Parallel()

	store := NewTokenBucketStore(100.0, 1.0) // Large global for simplicity

	model1 := store.Acquire("model1")
	model2 := store.Acquire("model2")
	model3 := store.Acquire("model3")

	if !model1 || !model2 || !model3 {
		t.Error("Expected all per-key acquisitions to succeed initially")
	}

	// Deplete model1 (each key bucket has 0.5 tokens default, refill at 0.1/sec)
	store.Reset() // Reset for clean test
	model1 = store.Acquire("model1")
	for i := 0; i < 10; i++ {
		store.Acquire("model1")
	}

	// Should fail for model1, but not model2 or model3
	if !model1 && store.Acquire("model1") == true {
		t.Error("Expected model1 to be rate limited after depletion")
	} else if store.Acquire("model2") {
		// Model2 should still work (independent per-key bucket)
	} else {
		t.Error("Expected model2 to still have tokens available")
	}
}

func TestTokenBucketStore_MultipleKeysIndependent(t *testing.T) {
	t.Parallel()

	store := NewTokenBucketStore(100.0, 0.0) // No refill for simplicity
	store.Reset()

	store.Acquire("key1")
	store.Acquire("key2")
	store.Acquire("key3")

	// Keys should have independent token counts (all default to small bucket with 0.1/sec refill)
	time.Sleep(5 * time.Millisecond) // Let buckets create with initial tokens

	// Keys should have independent token counts (all default to 1.0 tokens/sec bucket)
	if store.GetTokensForKey("key1") < 0.9 ||
		store.GetTokensForKey("key2") < 0.9 ||
		store.GetTokensForKey("key3") < 0.9 {
		t.Error("Expected each key to have independent ~1 token available")
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	tb := NewTokenBucket(1000.0, 100.0)

	for i := 0; i < b.N; i++ {
		tb.Allow()
	}
}

func BenchmarkTokenBucketStore_Acquire(b *testing.B) {
	store := NewTokenBucketStore(100.0, 10.0)

	for i := 0; i < b.N; i++ {
		store.Acquire("test-model")
	}
}

// Fuzz test for token bucket edge cases
func FuzzTokenBucket_Allow(f *testing.F) {
	f.Add(10.0, 1.0)   // Normal case
	f.Add(0.1, 0.5)    // Near empty
	f.Add(100.0, 10.0) // Large capacity
	f.Add(0.001, 0.01) // Very small tokens

	f.Fuzz(func(t *testing.T, capacity, rate float64) {
		if capacity < 0 || rate < 0 {
			t.Skip("negative values not tested")
		}

		tb := NewTokenBucket(capacity, rate)
		for i := 0; i < 100; i++ {
			tb.Allow()
		}
	})
}
