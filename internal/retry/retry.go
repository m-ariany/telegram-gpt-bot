package retry

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"
	"time"
)

type RetryHandler struct {
	rndMu               sync.Mutex
	maxDelay, maxJitter time.Duration
	maxRetry            int
}

func NewRetryHandler(MaxDelay, MaxJitter time.Duration, maxRetry int) *RetryHandler {
	return &RetryHandler{
		rndMu:     sync.Mutex{},
		maxDelay:  MaxDelay,
		maxJitter: MaxJitter,
		maxRetry:  maxRetry,
	}
}

type CallFunc func() error

func (b *RetryHandler) Do(c CallFunc) {
	for i := 1; i < b.maxRetry+1; i++ {
		if err := c(); err == nil {
			return
		}
		b.backoff(i)
	}
}

// Backoff is blocking and will return after the backoff duration.
func (b *RetryHandler) backoff(retryCount int) {

	if b.maxDelay == 0 {
		b.maxDelay = 5000 * time.Millisecond
	}
	if b.maxJitter == 0 {
		b.maxJitter = 2000 * time.Millisecond
	}

	b.rndMu.Lock()
	defer b.rndMu.Unlock()

	t := time.Duration(1<<uint(retryCount)) * time.Second
	backoff := time.Duration(math.Min(float64(t), float64(b.maxDelay)))
	center := backoff / 2
	var ri = int64(center)
	var jitter = newRnd(ri)

	sleepTime := time.Duration(math.Abs(float64(ri + jitter)))
	if sleepTime > b.maxDelay {
		sleepTime = b.maxDelay
	}
	<-time.After(sleepTime)
}

func newRnd(cap int64) int64 {
	// Generate a random number between 0 and cap
	randomInt, err := rand.Int(rand.Reader, big.NewInt(cap-1))
	if err != nil {
		return 0
	}

	return randomInt.Int64()
}
