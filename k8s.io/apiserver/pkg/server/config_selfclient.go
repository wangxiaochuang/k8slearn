package server

import (
	"fmt"
	"net"

	restclient "k8s.io/client-go/rest"
	netutils "k8s.io/utils/net"
)

const LoopbackClientServerNameOverride = "apiserver-loopback-client"

func (s *SecureServingInfo) NewClientConfig(caCert []byte) (*restclient.Config, error) {
	if s == nil || (s.Cert == nil && len(s.SNICerts) == 0) {
		return nil, nil
	}

	host, port, err := LoopbackHostPort(s.Listener.Addr().String())
	if err != nil {
		return nil, err
	}

	return &restclient.Config{
		// Do not limit loopback client QPS.
		QPS:  -1,
		Host: "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: restclient.TLSClientConfig{
			CAData: caCert,
		},
	}, nil
}

func (s *SecureServingInfo) NewLoopbackClientConfig(token string, loopbackCert []byte) (*restclient.Config, error) {
	c, err := s.NewClientConfig(loopbackCert)
	if err != nil || c == nil {
		return c, err
	}

	c.BearerToken = token
	c.TLSClientConfig.ServerName = LoopbackClientServerNameOverride

	return c, nil
}

func LoopbackHostPort(bindAddress string) (string, string, error) {
	host, port, err := net.SplitHostPort(bindAddress)
	if err != nil {
		return "", "", fmt.Errorf("invalid server bind address: %q", bindAddress)
	}

	isIPv6 := netutils.IsIPv6String(host)

	// Value is expected to be an IP or DNS name, not "0.0.0.0".
	if host == "0.0.0.0" || host == "::" {
		host = getLoopbackAddress(isIPv6)
	}
	return host, port, nil
}

func getLoopbackAddress(wantIPv6 bool) string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsLoopback() && wantIPv6 == netutils.IsIPv6(ipnet.IP) {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
