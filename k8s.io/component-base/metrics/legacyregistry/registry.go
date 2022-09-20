package legacyregistry

import "k8s.io/component-base/metrics"

var (
	defaultRegistry = metrics.NewKubeRegistry()
	// DefaultGatherer exposes the global registry gatherer
	DefaultGatherer metrics.Gatherer = defaultRegistry
	// Reset calls reset on the global registry
	Reset = defaultRegistry.Reset
	// 默认的registry
	MustRegister = defaultRegistry.MustRegister
	// RawMustRegister registers prometheus collectors but uses the global registry, this
	// bypasses the metric stability framework
	//
	// Deprecated
	RawMustRegister = defaultRegistry.RawMustRegister

	// Register registers a collectable metric but uses the global registry
	Register = defaultRegistry.Register
)
