package webhook

import (
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultRequestTimeout = 30 * time.Second

func DefaultRetryBackoffWithInitialDelay(initialBackoffDelay time.Duration) wait.Backoff {
	return wait.Backoff{
		Duration: initialBackoffDelay,
		Factor:   1.5,
		Jitter:   0.2,
		Steps:    5,
	}
}

func LoadKubeconfig(kubeConfigFile string, customDial utilnet.DialFunc) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeConfigFile
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	clientConfig, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientConfig.Dial = customDial

	clientConfig.Timeout = defaultRequestTimeout

	clientConfig.QPS = -1

	return clientConfig, nil
}
