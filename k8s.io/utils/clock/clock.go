package clock

import "time"

type PassiveClock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type Clock interface {
	PassiveClock
	After(d time.Duration) <-chan time.Time
	NewTimer(d time.Duration) Timer
	Sleep(d time.Duration)
	Tick(d time.Duration) <-chan time.Time
}

type WithTicker interface {
	Clock
	NewTicker(time.Duration) Ticker
}

type WithDelayedExecution interface {
	Clock
	AfterFunc(d time.Duration, f func()) Timer
}

type WithTickerAndDelayedExecution interface {
	WithTicker
	AfterFunc(d time.Duration, f func()) Timer
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

var _ = WithTicker(RealClock{})

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) Since(ts time.Time) time.Duration {
	return time.Since(ts)
}

func (RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (RealClock) NewTimer(d time.Duration) Timer {
	return &realTimer{
		timer: time.NewTimer(d),
	}
}

func (RealClock) AfterFunc(d time.Duration, f func()) Timer {
	return &realTimer{
		timer: time.AfterFunc(d, f),
	}
}

func (RealClock) Tick(d time.Duration) <-chan time.Time {
	return time.Tick(d)
}

func (RealClock) NewTicker(d time.Duration) Ticker {
	return &realTicker{
		ticker: time.NewTicker(d),
	}
}

func (RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

var _ = Timer(&realTimer{})

type realTimer struct {
	timer *time.Timer
}

func (r *realTimer) C() <-chan time.Time {
	return r.timer.C
}

func (r *realTimer) Stop() bool {
	return r.timer.Stop()
}

func (r *realTimer) Reset(d time.Duration) bool {
	return r.timer.Reset(d)
}

type realTicker struct {
	ticker *time.Ticker
}

func (r *realTicker) C() <-chan time.Time {
	return r.ticker.C
}

func (r *realTicker) Stop() {
	r.ticker.Stop()
}
