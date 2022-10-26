package webhook

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/x509metrics"
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

type GenericWebhook struct {
	RestClient   *rest.RESTClient
	RetryBackoff wait.Backoff
	ShouldRetry  func(error) bool
}

func DefaultShouldRetry(err error) bool {
	// these errors indicate a transient error that should be retried.
	if utilnet.IsConnectionReset(err) || apierrors.IsInternalError(err) || apierrors.IsTimeout(err) || apierrors.IsTooManyRequests(err) {
		return true
	}
	// if the error sends the Retry-After header, we respect it as an explicit confirmation we should retry.
	if _, shouldRetry := apierrors.SuggestsClientDelay(err); shouldRetry {
		return true
	}
	return false
}

func NewGenericWebhook(scheme *runtime.Scheme, codecFactory serializer.CodecFactory, config *rest.Config, groupVersions []schema.GroupVersion, retryBackoff wait.Backoff) (*GenericWebhook, error) {
	for _, groupVersion := range groupVersions {
		if !scheme.IsVersionRegistered(groupVersion) {
			return nil, fmt.Errorf("webhook plugin requires enabling extension resource: %s", groupVersion)
		}
	}

	clientConfig := rest.CopyConfig(config)

	codec := codecFactory.LegacyCodec(groupVersions...)
	clientConfig.ContentConfig.NegotiatedSerializer = serializer.NegotiatedSerializerWrapper(runtime.SerializerInfo{Serializer: codec})

	clientConfig.Wrap(x509metrics.NewDeprecatedCertificateRoundTripperWrapperConstructor(
		x509MissingSANCounter,
		x509InsecureSHA1Counter,
	))

	restClient, err := rest.UnversionedRESTClientFor(clientConfig)
	if err != nil {
		return nil, err
	}

	return &GenericWebhook{restClient, retryBackoff, DefaultShouldRetry}, nil
}

func (g *GenericWebhook) WithExponentialBackoff(ctx context.Context, webhookFn func() rest.Result) rest.Result {
	var result rest.Result
	shouldRetry := g.ShouldRetry
	if shouldRetry == nil {
		shouldRetry = DefaultShouldRetry
	}
	WithExponentialBackoff(ctx, g.RetryBackoff, func() error {
		result = webhookFn()
		return result.Error()
	}, shouldRetry)
	return result
}

func WithExponentialBackoff(ctx context.Context, retryBackoff wait.Backoff, webhookFn func() error, shouldRetry func(error) bool) error {
	// having a webhook error allows us to track the last actual webhook error for requests that
	// are later cancelled or time out.
	var webhookErr error
	err := wait.ExponentialBackoffWithContext(ctx, retryBackoff, func() (bool, error) {
		webhookErr = webhookFn()
		if shouldRetry(webhookErr) {
			return false, nil
		}
		if webhookErr != nil {
			return false, webhookErr
		}
		return true, nil
	})

	switch {
	// we check for webhookErr first, if webhookErr is set it's the most important error to return.
	case webhookErr != nil:
		return webhookErr
	case err != nil:
		return fmt.Errorf("webhook call failed: %s", err.Error())
	default:
		return nil
	}
}

// p146
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
