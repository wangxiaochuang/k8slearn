package rest

import (
	"fmt"
	"net/http"
	"sync"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

type AuthProvider interface {
	WrapTransport(http.RoundTripper) http.RoundTripper
	Login() error
}

type Factory func(clusterAddress string, config map[string]string, persister AuthProviderConfigPersister) (AuthProvider, error)

type noopPersister struct{}

func (n *noopPersister) Persist(_ map[string]string) error {
	// no operation persister
	return nil
}

var pluginsLock sync.Mutex
var plugins = make(map[string]Factory)

func RegisterAuthProviderPlugin(name string, plugin Factory) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	if _, found := plugins[name]; found {
		return fmt.Errorf("auth Provider Plugin %q was registered twice", name)
	}
	klog.V(4).Infof("Registered Auth Provider Plugin %q", name)
	plugins[name] = plugin
	return nil
}

type AuthProviderConfigPersister interface {
	Persist(map[string]string) error
}

func GetAuthProvider(clusterAddress string, apc *clientcmdapi.AuthProviderConfig, persister AuthProviderConfigPersister) (AuthProvider, error) {
	panic("not implemented")
}
