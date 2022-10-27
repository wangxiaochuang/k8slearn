package reconcilers

import (
	"net"

	corev1 "k8s.io/api/core/v1"
)

type EndpointReconciler interface {
	ReconcileEndpoints(serviceName string, ip net.IP, endpointPorts []corev1.EndpointPort, reconcilePorts bool) error
	RemoveEndpoints(serviceName string, ip net.IP, endpointPorts []corev1.EndpointPort) error
	StopReconciling()
}

type Type string

const (
	MasterCountReconcilerType   Type = "master-count"
	LeaseEndpointReconcilerType      = "lease"
	NoneEndpointReconcilerType       = "none"
)

type Types []Type

var AllTypes = Types{
	MasterCountReconcilerType,
	LeaseEndpointReconcilerType,
	NoneEndpointReconcilerType,
}

func (t Types) Names() []string {
	strs := make([]string, len(t))
	for i, v := range t {
		strs[i] = string(v)
	}
	return strs
}
