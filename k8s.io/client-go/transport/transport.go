package transport

import (
	"crypto/tls"
	"net/http"
)

func New(config *Config) (http.RoundTripper, error) {
	panic("not implemented")
}

// p61
func TLSConfigFor(c *Config) (*tls.Config, error) {
	panic("not implemented")
}

// p224
type WrapperFunc func(rt http.RoundTripper) http.RoundTripper

func Wrappers(fns ...WrapperFunc) WrapperFunc {
	if len(fns) == 0 {
		return nil
	}
	if len(fns) == 2 && fns[0] == nil {
		return fns[1]
	}
	return func(rt http.RoundTripper) http.RoundTripper {
		base := rt
		for _, fn := range fns {
			if fn != nil {
				base = fn(base)
			}
		}
		return base
	}
}
