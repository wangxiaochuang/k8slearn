package leaderelection

import "sync"

type leaderMetricsAdapter interface {
	leaderOn(name string)
	leaderOff(name string)
}

type SwitchMetric interface {
	On(name string)
	Off(name string)
}

type noopMetric struct{}

func (noopMetric) On(name string)  {}
func (noopMetric) Off(name string) {}

type defaultLeaderMetrics struct {
	leader SwitchMetric
}

func (m *defaultLeaderMetrics) leaderOn(name string) {
	if m == nil {
		return
	}
	m.leader.On(name)
}

func (m *defaultLeaderMetrics) leaderOff(name string) {
	if m == nil {
		return
	}
	m.leader.Off(name)
}

type noMetrics struct{}

func (noMetrics) leaderOn(name string)  {}
func (noMetrics) leaderOff(name string) {}

type MetricsProvider interface {
	NewLeaderMetric() SwitchMetric
}

type noopMetricsProvider struct{}

func (_ noopMetricsProvider) NewLeaderMetric() SwitchMetric {
	return noopMetric{}
}

var globalMetricsFactory = leaderMetricsFactory{
	metricsProvider: noopMetricsProvider{},
}

type leaderMetricsFactory struct {
	metricsProvider MetricsProvider

	onlyOnce sync.Once
}

func (f *leaderMetricsFactory) setProvider(mp MetricsProvider) {
	f.onlyOnce.Do(func() {
		f.metricsProvider = mp
	})
}

func (f *leaderMetricsFactory) newLeaderMetrics() leaderMetricsAdapter {
	mp := f.metricsProvider
	if mp == (noopMetricsProvider{}) {
		return noMetrics{}
	}
	return &defaultLeaderMetrics{
		leader: mp.NewLeaderMetric(),
	}
}

func SetProvider(metricsProvider MetricsProvider) {
	globalMetricsFactory.setProvider(metricsProvider)
}
