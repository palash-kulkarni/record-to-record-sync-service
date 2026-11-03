package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// ProviderRateConfig holds rate limit settings for a single CRM provider.
type ProviderRateConfig struct {
	RequestsPerSecond int
	BurstCapacity     int
}

// tokenBucket is a per-(provider, accountID) rate limiter.
// Token bucket allows short bursts up to BurstCapacity while enforcing the long-term average rate.
type tokenBucket struct {
	mu             sync.Mutex
	capacity       float64
	refillPerSec   float64
	tokens         float64
	lastRefillTime time.Time
}

func newTokenBucket(requestsPerSecond, burstCapacity int) *tokenBucket {
	return &tokenBucket{
		capacity:       float64(burstCapacity),
		refillPerSec:   float64(requestsPerSecond),
		tokens:         float64(burstCapacity), // start full so the first burst is available immediately
		lastRefillTime: time.Now(),
	}
}

func (b *tokenBucket) tryConsume() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefillTime)

	// Add tokens proportional to elapsed time, then cap at capacity.
	// We do this lazily on each call rather than in a background goroutine.
	b.tokens += elapsed.Seconds() * b.refillPerSec
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefillTime = now

	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter manages one token bucket per (providerID, accountID) pair.
// Buckets are created lazily on first use. In a multi-instance deployment,
// this state would need to move to Redis so all instances share the same quota.
type RateLimiter struct {
	configs map[string]ProviderRateConfig
	buckets sync.Map // key: "providerID:accountID"
}

func NewRateLimiter(configs map[string]ProviderRateConfig) *RateLimiter {
	return &RateLimiter{configs: configs}
}

// TryAcquire attempts to consume one token for the given provider and account.
// Returns true if allowed, false if rate-limited.
func (r *RateLimiter) TryAcquire(providerID, accountID string) bool {
	config, ok := r.configs[providerID]
	if !ok {
		// Unknown provider, allow through and let provider handle it
		return true
	}

	key := fmt.Sprintf("%s:%s", providerID, accountID)
	bucket, _ := r.buckets.LoadOrStore(key, newTokenBucket(config.RequestsPerSecond, config.BurstCapacity))
	return bucket.(*tokenBucket).tryConsume()
}
