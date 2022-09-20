package options

import (
	genericoptions "k8s.io/apiserver/pkg/server/options"
	netutils "k8s.io/utils/net"
)

func NewSecureServingOptions() *genericoptions.SecureServingOptionsWithLoopback {
	o := genericoptions.SecureServingOptions{
		BindAddress: netutils.ParseIPSloppy("0.0.0.0"),
		BindPort:    6443,
		Required:    true,
		ServerCert: genericoptions.GeneratableKeyCert{
			PairName:      "apiserver",
			CertDirectory: "/var/run/kubernetes",
		},
	}
	return o.WithLoopback()
}
