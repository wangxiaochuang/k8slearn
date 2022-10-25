package clientcmd

import (
	"io"
	"sync"

	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type DeferredLoadingClientConfig struct {
	loader         ClientConfigLoader
	overrides      *ConfigOverrides
	fallbackReader io.Reader

	clientConfig ClientConfig
	loadingLock  sync.Mutex

	// provided for testing
	icc InClusterConfig
}

type InClusterConfig interface {
	ClientConfig
	Possible() bool
}

func NewNonInteractiveDeferredLoadingClientConfig(loader ClientConfigLoader, overrides *ConfigOverrides) ClientConfig {
	return &DeferredLoadingClientConfig{loader: loader, overrides: overrides, icc: &inClusterClientConfig{overrides: overrides}}
}

func NewInteractiveDeferredLoadingClientConfig(loader ClientConfigLoader, overrides *ConfigOverrides, fallbackReader io.Reader) ClientConfig {
	return &DeferredLoadingClientConfig{loader: loader, overrides: overrides, icc: &inClusterClientConfig{overrides: overrides}, fallbackReader: fallbackReader}
}

func (config *DeferredLoadingClientConfig) createClientConfig() (ClientConfig, error) {
	panic("not implemented")
}

func (config *DeferredLoadingClientConfig) RawConfig() (clientcmdapi.Config, error) {
	mergedConfig, err := config.createClientConfig()
	if err != nil {
		return clientcmdapi.Config{}, err
	}

	return mergedConfig.RawConfig()
}

func (config *DeferredLoadingClientConfig) ClientConfig() (*restclient.Config, error) {
	panic("not implemented")
}

// Namespace implements KubeConfig
func (config *DeferredLoadingClientConfig) Namespace() (string, bool, error) {
	panic("not implemented")
}

// ConfigAccess implements ClientConfig
func (config *DeferredLoadingClientConfig) ConfigAccess() ConfigAccess {
	return config.loader
}
