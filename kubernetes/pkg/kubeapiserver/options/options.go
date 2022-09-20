package options

import (
	"net"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	netutils "k8s.io/utils/net"
)

var DefaultServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}

var DefaultServiceIPCIDR = net.IPNet{IP: netutils.ParseIPSloppy("10.0.0.0"), Mask: net.CIDRMask(24, 32)}

const DefaultEtcdPathPrefix = "/registry"
