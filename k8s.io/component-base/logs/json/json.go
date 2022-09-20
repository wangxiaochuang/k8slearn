package json

import (
	"io"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs/registry"
)

type Factory struct{}

var _ registry.LogFormatFactory = Factory{}

func (f Factory) Create(c config.LoggingConfiguration) (logr.Logger, func()) {
	panic("not impl")
}

func AddNopSync(writer io.Writer) zapcore.WriteSyncer {
	return nopSync{Writer: writer}
}

type nopSync struct {
	io.Writer
}

func (f nopSync) Sync() error {
	return nil
}
