package throttle

import (
	"math"
	"sync"
	"time"
)

type TokenBucket struct {
	id           string
	maxAmount    int64
	refillTime   time.Duration
	refillAmount int64
	currValue    int64
	lastRefill   time.Time
	mutex        *sync.Mutex
}

func NewTokenBucket(key string, maxAmount int64, refillTime time.Duration, refillAmount int64) *TokenBucket {
	return &TokenBucket{
		id:           key,
		maxAmount:    maxAmount,
		refillTime:   refillTime,
		refillAmount: refillAmount,
		currValue:    maxAmount,
		lastRefill:   time.Now(),
		mutex:        &sync.Mutex{},
	}
}

func (t *TokenBucket) GetThrottleDecision() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.replenishTokens()
	if t.currValue >= 1 {
		t.currValue--
		return false
	}
	return true
}

func (t *TokenBucket) replenishTokens() {
	now := time.Now()
	timeDelta := now.Sub(t.lastRefill)
	if timeDelta.Seconds() < t.refillTime.Seconds() {
		return
	}
	refillIntervalsSinceLastUpdate := int64(math.Floor(timeDelta.Seconds() / t.refillTime.Seconds()))
	newTokenCount := t.currValue + (refillIntervalsSinceLastUpdate * t.refillAmount)
	t.currValue = min(t.maxAmount, newTokenCount)
	t.lastRefill = now
}

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
