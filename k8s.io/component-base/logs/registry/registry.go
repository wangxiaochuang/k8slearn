package registry

import (
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	"k8s.io/component-base/config"
)

var LogRegistry = NewLogFormatRegistry()

type LogFormatRegistry struct {
	registry map[string]LogFormatFactory
	frozen   bool
}

type LogFormatFactory interface {
	Create(c config.LoggingConfiguration) (log logr.Logger, flush func())
}

func NewLogFormatRegistry() *LogFormatRegistry {
	return &LogFormatRegistry{
		registry: make(map[string]LogFormatFactory),
		frozen:   false,
	}
}

func (lfr *LogFormatRegistry) Register(name string, factory LogFormatFactory) error {
	// frozen后，就不能再调用注册函数了
	if lfr.frozen {
		return fmt.Errorf("log format is frozen, unable to register log format")
	}
	// 不能重复注册
	if _, ok := lfr.registry[name]; ok {
		return fmt.Errorf("log format: %s already exists", name)
	}
	lfr.registry[name] = factory
	return nil
}

func (lfr *LogFormatRegistry) Get(name string) (LogFormatFactory, error) {
	re, ok := lfr.registry[name]
	if !ok {
		return nil, fmt.Errorf("log format: %s does not exists", name)
	}
	return re, nil
}

func (lfr *LogFormatRegistry) Set(name string, factory LogFormatFactory) error {
	if lfr.frozen {
		return fmt.Errorf("log format is frozen, unable to set log format")
	}

	lfr.registry[name] = factory
	return nil
}

func (lfr *LogFormatRegistry) Delete(name string) error {
	if lfr.frozen {
		return fmt.Errorf("log format is frozen, unable to delete log format")
	}

	delete(lfr.registry, name)
	return nil
}

func (lfr *LogFormatRegistry) List() []string {
	formats := make([]string, 0, len(lfr.registry))
	for f := range lfr.registry {
		formats = append(formats, f)
	}
	sort.Strings(formats)
	return formats
}

func (lfr *LogFormatRegistry) Freeze() {
	lfr.frozen = true
}
