package apiserver

import (
	"net/url"

	listersv1 "k8s.io/client-go/listers/core/v1"
)

type ServiceResolver interface {
	ResolveEndpoint(namespace, name string, port int32) (*url.URL, error)
}

func NewEndpointServiceResolver(services listersv1.ServiceLister, endpoints listersv1.EndpointsLister) ServiceResolver {
	return &aggregatorEndpointRouting{
		services:  services,
		endpoints: endpoints,
	}
}

type aggregatorEndpointRouting struct {
	services  listersv1.ServiceLister
	endpoints listersv1.EndpointsLister
}

func (r *aggregatorEndpointRouting) ResolveEndpoint(namespace, name string, port int32) (*url.URL, error) {
	panic("not implemented")
}

func NewClusterIPServiceResolver(services listersv1.ServiceLister) ServiceResolver {
	return &aggregatorClusterRouting{
		services: services,
	}
}

type aggregatorClusterRouting struct {
	services listersv1.ServiceLister
}

func (r *aggregatorClusterRouting) ResolveEndpoint(namespace, name string, port int32) (*url.URL, error) {
	panic("not implemented")
}

func NewLoopbackServiceResolver(delegate ServiceResolver, host *url.URL) ServiceResolver {
	return &loopbackResolver{
		delegate: delegate,
		host:     host,
	}
}

type loopbackResolver struct {
	delegate ServiceResolver
	host     *url.URL
}

func (r *loopbackResolver) ResolveEndpoint(namespace, name string, port int32) (*url.URL, error) {
	if namespace == "default" && name == "kubernetes" && port == 443 {
		return r.host, nil
	}
	return r.delegate.ResolveEndpoint(namespace, name, port)
}
