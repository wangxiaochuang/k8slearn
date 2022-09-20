package apiserver

import "net/url"

type ServiceResolver interface {
	ResolveEndpoint(namespace, name string, port int32) (*url.URL, error)
}
