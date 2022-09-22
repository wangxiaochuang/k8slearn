package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EgressSelectorConfiguration struct {
	metav1.TypeMeta  `json:",inline"`
	EgressSelections []EgressSelection `json:"egressSelections"`
}

type EgressSelection struct {
	Name       string     `json:"name"`
	Connection Connection `json:"connection"`
}

type Connection struct {
	ProxyProtocol ProtocolType `json:"proxyProtocol,omitempty"`
	Transport     *Transport   `json:"transport,omitempty"`
}

type ProtocolType string

const (
	ProtocolHTTPConnect ProtocolType = "HTTPConnect"
	ProtocolGRPC        ProtocolType = "GRPC"
	ProtocolDirect      ProtocolType = "Direct"
)

// Transport defines the transport configurations we use to dial to the konnectivity server
type Transport struct {
	TCP *TCPTransport `json:"tcp,omitempty"`
	UDS *UDSTransport `json:"uds,omitempty"`
}

type TCPTransport struct {
	URL       string     `json:"url,omitempty"`
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
}

type UDSTransport struct {
	UDSName string `json:"udsName,omitempty"`
}

type TLSConfig struct {
	CABundle   string `json:"caBundle,omitempty"`
	ClientKey  string `json:"clientKey,omitempty"`
	ClientCert string `json:"clientCert,omitempty"`
}
