package ratelimit_test

import (
	"testing"
	"time"

	"github.com/palash-kulkarni/record-to-record-sync-service/ratelimit"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 10, BurstCapacity: 5},
	})

	for i := 0; i < 5; i++ {
		if !rl.TryAcquire("salesforce", "acct-1") {
			t.Fatalf("call %d should be allowed within burst capacity", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	rl := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 1, BurstCapacity: 2},
	})

	rl.TryAcquire("salesforce", "acct-1")
	rl.TryAcquire("salesforce", "acct-1")

	if rl.TryAcquire("salesforce", "acct-1") {
		t.Error("third call should be rate-limited after burst exhausted")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 100, BurstCapacity: 1},
	})

	rl.TryAcquire("salesforce", "acct-1") // exhaust
	time.Sleep(20 * time.Millisecond)      // wait for refill (~2 tokens at 100 RPS)

	if !rl.TryAcquire("salesforce", "acct-1") {
		t.Error("should be allowed after refill")
	}
}

func TestRateLimiter_IsolatesAccounts(t *testing.T) {
	rl := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 1, BurstCapacity: 1},
	})

	rl.TryAcquire("salesforce", "acct-1") // exhaust acct-1

	// acct-2 should still be allowed, it has a separate bucket
	if !rl.TryAcquire("salesforce", "acct-2") {
		t.Error("acct-2 should not be affected by acct-1 exhaustion")
	}
}

func TestRateLimiter_UnknownProviderAllowed(t *testing.T) {
	rl := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{})

	if !rl.TryAcquire("zoho", "acct-1") {
		t.Error("unknown provider should be allowed through")
	}
}
