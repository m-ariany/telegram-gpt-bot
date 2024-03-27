package limiter

import (
	"context"
	"testing"

	"github.com/go-redis/redis_rate/v10"
	"github.com/stretchr/testify/assert"
)

// MockLimiterBackend is a mock type for the LimiterBackend
type MockLimiterBackend struct {
}

func (m *MockLimiterBackend) AllowN(ctx context.Context, key string, limit redis_rate.Limit, n int) (*redis_rate.Result, error) {
	return nil, nil
}

func (m *MockLimiterBackend) Reset(ctx context.Context, key string) error {
	return nil
}

func TestRedisLimiterRate(t *testing.T) {
	mockLimiter := new(MockLimiterBackend)
	rl := NewLimiter(mockLimiter, "testPrefix")

	// Testing PerSecond
	limiter := rl.PerSecond(10)
	assert.Equal(t, 10, limiter.lb.limit.Burst)
	assert.Equal(t, "testPrefix-second", limiter.lb.keyPrefix)

	// Testing PerMinute
	rl = NewLimiter(mockLimiter, "testPrefix")
	limiter = rl.PerMinute(20)
	assert.Equal(t, 20, limiter.lb.limit.Burst)
	assert.Equal(t, "testPrefix-minute", limiter.lb.keyPrefix)

	// Testing PerHour
	rl = NewLimiter(mockLimiter, "testPrefix")
	limiter = rl.PerHour(30)
	assert.Equal(t, 30, limiter.lb.limit.Burst)
	assert.Equal(t, "testPrefix-hour", limiter.lb.keyPrefix)

	// Testing PerDay
	rl = NewLimiter(mockLimiter, "testPrefix")
	limiter = rl.PerDay(40)
	assert.Equal(t, 40, limiter.lb.limit.Burst)
	assert.Equal(t, "testPrefix-day", limiter.lb.keyPrefix)
}
