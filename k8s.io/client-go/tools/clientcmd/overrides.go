package clientcmd

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ConfigOverrides struct {
	AuthInfo clientcmdapi.AuthInfo
	// ClusterDefaults are applied before the configured cluster info is loaded.
	ClusterDefaults clientcmdapi.Cluster
	ClusterInfo     clientcmdapi.Cluster
	Context         clientcmdapi.Context
	CurrentContext  string
	Timeout         string
}
