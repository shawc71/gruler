package throttle

import (
	"sync"
	"time"
)

type BucketManager struct {
	buckets map[string]*TokenBucket
	mutex   *sync.Mutex
}

func NewBucketManager() *BucketManager {
	buckets := make(map[string]*TokenBucket)
	return &BucketManager{
		buckets: buckets,
		mutex:   &sync.Mutex{},
	}
}

func (b *BucketManager) GetBucket(key string, maxAmount int64, refillTime time.Duration, refillAmount int64) *TokenBucket {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, exists := b.buckets[key]
	if !exists {
		bucket := NewTokenBucket(key, maxAmount, refillTime, refillAmount)
		b.buckets[key] = bucket
	}
	return b.buckets[key]
}
