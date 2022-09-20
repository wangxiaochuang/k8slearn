package discovery

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Addresses interface {
	ServerAddressByClientCIDRs(net.IP) []metav1.ServerAddressByClientCIDR
}
