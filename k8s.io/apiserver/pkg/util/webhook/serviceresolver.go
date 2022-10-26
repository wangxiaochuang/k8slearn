package webhook

import (
	"errors"
	"fmt"
	"net/url"
)

// ServiceResolver knows how to convert a service reference into an actual location.
type ServiceResolver interface {
	ResolveEndpoint(namespace, name string, port int32) (*url.URL, error)
}

type defaultServiceResolver struct{}

// NewDefaultServiceResolver creates a new default server resolver.
func NewDefaultServiceResolver() ServiceResolver {
	return &defaultServiceResolver{}
}

// ResolveEndpoint constructs a service URL from a given namespace and name
// note that the name, namespace, and port are required and by default all
// created addresses use HTTPS scheme.
// for example:
//
//	name=ross namespace=andromeda resolves to https://ross.andromeda.svc:443
func (sr defaultServiceResolver) ResolveEndpoint(namespace, name string, port int32) (*url.URL, error) {
	if len(name) == 0 || len(namespace) == 0 || port == 0 {
		return nil, errors.New("cannot resolve an empty service name or namespace or port")
	}
	return &url.URL{Scheme: "https", Host: fmt.Sprintf("%s.%s.svc:%d", name, namespace, port)}, nil
}
