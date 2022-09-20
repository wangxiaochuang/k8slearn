package capabilities

import "sync"

type Capabilities struct {
	AllowPrivileged                        bool
	PrivilegedSources                      PrivilegedSources
	PerConnectionBandwidthLimitBytesPerSec int64
}

type PrivilegedSources struct {
	HostNetworkSources []string
	HostPIDSources     []string
	HostIPCSources     []string
}

var capInstance struct {
	once         sync.Once
	lock         sync.Mutex
	capabilities *Capabilities
}

func Initialize(c Capabilities) {
	capInstance.once.Do(func() {
		capInstance.capabilities = &c
	})
}

func Setup(allowPrivileged bool, perConnectionBytesPerSec int64) {
	Initialize(Capabilities{
		AllowPrivileged:                        allowPrivileged,
		PerConnectionBandwidthLimitBytesPerSec: perConnectionBytesPerSec,
	})
}

func SetForTests(c Capabilities) {
	capInstance.lock.Lock()
	defer capInstance.lock.Unlock()
	capInstance.capabilities = &c
}

func Get() Capabilities {
	capInstance.lock.Lock()
	defer capInstance.lock.Unlock()
	// This check prevents clobbering of capabilities that might've been set via SetForTests
	if capInstance.capabilities == nil {
		Initialize(Capabilities{
			AllowPrivileged: false,
			PrivilegedSources: PrivilegedSources{
				HostNetworkSources: []string{},
				HostPIDSources:     []string{},
				HostIPCSources:     []string{},
			},
		})
	}
	return *capInstance.capabilities
}
