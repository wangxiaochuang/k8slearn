package rest

import (
	"net/http"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type AuthProvider interface {
	WrapTransport(http.RoundTripper) http.RoundTripper
	Login() error
}

type AuthProviderConfigPersister interface {
	Persist(map[string]string) error
}

func GetAuthProvider(clusterAddress string, apc *clientcmdapi.AuthProviderConfig, persister AuthProviderConfigPersister) (AuthProvider, error) {
	panic("not implemented")
}
