package clientcmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// p514
type ClientConfig interface {
	// RawConfig returns the merged result of all overrides
	RawConfig() (clientcmdapi.Config, error)
	// ClientConfig returns a complete client config
	ClientConfig() (*restclient.Config, error)
	// Namespace returns the namespace resulting from the merged
	// result of all overrides and a boolean indicating if it was
	// overridden
	Namespace() (string, bool, error)
	// ConfigAccess returns the rules for loading/persisting the config.
	ConfigAccess() ConfigAccess
}

type inClusterClientConfig struct {
	overrides               *ConfigOverrides
	inClusterConfigProvider func() (*restclient.Config, error)
}

var _ ClientConfig = &inClusterClientConfig{}

func (config *inClusterClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, fmt.Errorf("inCluster environment config doesn't support multiple clusters")
}

func (config *inClusterClientConfig) ClientConfig() (*restclient.Config, error) {
	panic("not implemented")
}

func (config *inClusterClientConfig) Namespace() (string, bool, error) {
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, false, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, false, nil
		}
	}

	return "default", false, nil
}

func (config *inClusterClientConfig) ConfigAccess() ConfigAccess {
	return NewDefaultClientConfigLoadingRules()
}

func (config *inClusterClientConfig) Possible() bool {
	fi, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != "" &&
		err == nil && !fi.IsDir()
}
