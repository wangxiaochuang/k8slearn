package flowcontrol

import "context"

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
