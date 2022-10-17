package flowcontrol

import (
	"context"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/utils/clock"
)

type PassiveRateLimiter interface {
	TryAccept() bool
	Stop()
	QPS() float32
}

type RateLimiter interface {
	PassiveRateLimiter
	Accept()
	Wait(ctx context.Context) error
}

type tokenBucketPassiveRateLimiter struct {
	limiter *rate.Limiter
	qps     float32
	clock   clock.PassiveClock
}

type tokenBucketRateLimiter struct {
	tokenBucketPassiveRateLimiter
	clock Clock
}

func NewTokenBucketRateLimiter(qps float32, burst int) RateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return newTokenBucketRateLimiterWithClock(limiter, clock.RealClock{}, qps)
}

func NewTokenBucketPassiveRateLimiter(qps float32, burst int) PassiveRateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return newTokenBucketRateLimiterWithPassiveClock(limiter, clock.RealClock{}, qps)
}

type Clock interface {
	clock.PassiveClock
	Sleep(time.Duration)
}

var _ Clock = (*clock.RealClock)(nil)

func NewTokenBucketRateLimiterWithClock(qps float32, burst int, c Clock) RateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return newTokenBucketRateLimiterWithClock(limiter, c, qps)
}

func NewTokenBucketPassiveRateLimiterWithClock(qps float32, burst int, c clock.PassiveClock) PassiveRateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return newTokenBucketRateLimiterWithPassiveClock(limiter, c, qps)
}

func newTokenBucketRateLimiterWithClock(limiter *rate.Limiter, c Clock, qps float32) *tokenBucketRateLimiter {
	return &tokenBucketRateLimiter{
		tokenBucketPassiveRateLimiter: *newTokenBucketRateLimiterWithPassiveClock(limiter, c, qps),
		clock:                         c,
	}
}

func newTokenBucketRateLimiterWithPassiveClock(limiter *rate.Limiter, c clock.PassiveClock, qps float32) *tokenBucketPassiveRateLimiter {
	return &tokenBucketPassiveRateLimiter{
		limiter: limiter,
		qps:     qps,
		clock:   c,
	}
}

func (tbprl *tokenBucketPassiveRateLimiter) Stop() {
}

func (tbprl *tokenBucketPassiveRateLimiter) QPS() float32 {
	return tbprl.qps
}

func (tbprl *tokenBucketPassiveRateLimiter) TryAccept() bool {
	return tbprl.limiter.AllowN(tbprl.clock.Now(), 1)
}

// Accept will block until a token becomes available
func (tbrl *tokenBucketRateLimiter) Accept() {
	now := tbrl.clock.Now()
	tbrl.clock.Sleep(tbrl.limiter.ReserveN(now, 1).DelayFrom(now))
}

func (tbrl *tokenBucketRateLimiter) Wait(ctx context.Context) error {
	return tbrl.limiter.Wait(ctx)
}
