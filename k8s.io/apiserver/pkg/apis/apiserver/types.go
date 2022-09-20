package apiserver

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AdmissionConfiguration struct {
	metav1.TypeMeta
	Plugins []AdmissionPluginConfiguration
}

type AdmissionPluginConfiguration struct {
	Name          string
	Path          string
	Configuration *runtime.Unknown
}

type EgressSelectorConfiguration struct {
	metav1.TypeMeta
	EgressSelections []EgressSelection
}

type EgressSelection struct {
	Name       string
	Connection Connection
}

type Connection struct {
	ProxyProtocol ProtocolType
	Transport     *Transport
}

type ProtocolType string

const (
	ProtocolHTTPConnect ProtocolType = "HTTPConnect"
	ProtocolGRPC        ProtocolType = "GRPC"
	ProtocolDirect      ProtocolType = "Direct"
)

type Transport struct {
	TCP *TCPTransport
	UDS *UDSTransport
}

type TCPTransport struct {
	URL       string
	TLSConfig *TLSConfig
}

type UDSTransport struct {
	UDSName string
}

type TLSConfig struct {
	CABundle   string
	ClientKey  string
	ClientCert string
}

type TracingConfiguration struct {
	metav1.TypeMeta
	Endpoint               *string
	SamplingRatePerMillion *int32
}
