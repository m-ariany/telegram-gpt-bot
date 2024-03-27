package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

type Period string

const (
	Second Period = "second"
	Minute        = "minute"
	Hour          = "hour"
	Day           = "day"
)

type RedisLimiter struct {
	lb *LimitBucket
}

type LimitBucket struct {
	limiter   LimiterBackend
	limit     redis_rate.Limit
	period    Period
	rate      int
	keyPrefix string
}

type ILimitBucket interface {
	PerSecond(rate int) Limiter
	PerMinute(rate int) Limiter
	PerHour(rate int) Limiter
	PerDay(rate int) Limiter
}

type Limiter interface {
	Allow(ctx context.Context, key string) (*redis_rate.Result, error)
}

type LimiterBackend interface {
	AllowN(ctx context.Context, key string, limit redis_rate.Limit, n int) (*redis_rate.Result, error)
	Reset(ctx context.Context, key string) error
}

type iRedisLimiter interface {
	Limiter
}

var _ iRedisLimiter = &RedisLimiter{}

func NewLimiter(l LimiterBackend, keyPrefix string) *LimitBucket {
	return &LimitBucket{
		limiter:   l,
		keyPrefix: keyPrefix,
	}
}

func (l *LimitBucket) PerSecond(rate int) *RedisLimiter {
	l.limit = redis_rate.Limit{
		Rate:   rate,
		Period: time.Second,
		Burst:  rate,
	}
	l.rate = rate
	l.period = Second
	l.keyPrefix = fmt.Sprintf("%s-%s", l.keyPrefix, l.period)
	return &RedisLimiter{
		lb: l,
	}
}

func (l *LimitBucket) PerMinute(rate int) *RedisLimiter {
	l.limit = redis_rate.Limit{
		Rate:   rate,
		Period: time.Minute,
		Burst:  rate,
	}
	l.rate = rate
	l.period = Minute
	l.keyPrefix = fmt.Sprintf("%s-%s", l.keyPrefix, l.period)
	return &RedisLimiter{
		lb: l,
	}
}

func (l *LimitBucket) PerHour(rate int) *RedisLimiter {
	l.limit = redis_rate.Limit{
		Rate:   rate,
		Period: time.Hour,
		Burst:  rate,
	}
	l.rate = rate
	l.period = Hour
	l.keyPrefix = fmt.Sprintf("%s-%s", l.keyPrefix, l.period)
	return &RedisLimiter{
		lb: l,
	}
}

func (l *LimitBucket) PerDay(rate int) *RedisLimiter {
	l.limit = redis_rate.Limit{
		Rate:   rate,
		Period: time.Hour * 24,
		Burst:  rate,
	}
	l.rate = rate
	l.period = Day
	l.keyPrefix = fmt.Sprintf("%s-%s", l.keyPrefix, l.period)
	return &RedisLimiter{
		lb: l,
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) (*redis_rate.Result, error) {
	return l.lb.limiter.AllowN(ctx, l.key(key), l.lb.limit, 1)
}

func (l *RedisLimiter) key(s string) string {
	return fmt.Sprintf("%s-%s", l.lb.keyPrefix, s)
}
