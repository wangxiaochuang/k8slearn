package webhook

import (
	"context"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/util/webhook"
	authenticationv1client "k8s.io/client-go/kubernetes/typed/authentication/v1"
	"k8s.io/client-go/rest"
)

func DefaultRetryBackoff() *wait.Backoff {
	backoff := webhook.DefaultRetryBackoffWithInitialDelay(500 * time.Millisecond)
	return &backoff
}

// Ensure WebhookTokenAuthenticator implements the authenticator.Token interface.
var _ authenticator.Token = (*WebhookTokenAuthenticator)(nil)

type tokenReviewer interface {
	Create(ctx context.Context, review *authenticationv1.TokenReview, _ metav1.CreateOptions) (*authenticationv1.TokenReview, int, error)
}

type WebhookTokenAuthenticator struct {
	tokenReview    tokenReviewer
	retryBackoff   wait.Backoff
	implicitAuds   authenticator.Audiences
	requestTimeout time.Duration
	metrics        AuthenticatorMetrics
}

func NewFromInterface(tokenReview authenticationv1client.AuthenticationV1Interface, implicitAuds authenticator.Audiences, retryBackoff wait.Backoff, requestTimeout time.Duration, metrics AuthenticatorMetrics) (*WebhookTokenAuthenticator, error) {
	tokenReviewClient := &tokenReviewV1Client{tokenReview.RESTClient()}
	return newWithBackoff(tokenReviewClient, retryBackoff, implicitAuds, requestTimeout, metrics)
}

func New(config *rest.Config, version string, implicitAuds authenticator.Audiences, retryBackoff wait.Backoff) (*WebhookTokenAuthenticator, error) {
	tokenReview, err := tokenReviewInterfaceFromConfig(config, version, retryBackoff)
	if err != nil {
		return nil, err
	}
	return newWithBackoff(tokenReview, retryBackoff, implicitAuds, time.Duration(0), AuthenticatorMetrics{
		RecordRequestTotal:   noopMetrics{}.RequestTotal,
		RecordRequestLatency: noopMetrics{}.RequestLatency,
	})
}

func newWithBackoff(tokenReview tokenReviewer, retryBackoff wait.Backoff, implicitAuds authenticator.Audiences, requestTimeout time.Duration, metrics AuthenticatorMetrics) (*WebhookTokenAuthenticator, error) {
	return &WebhookTokenAuthenticator{
		tokenReview,
		retryBackoff,
		implicitAuds,
		requestTimeout,
		metrics,
	}, nil
}

func (w *WebhookTokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}

func tokenReviewInterfaceFromConfig(config *rest.Config, version string, retryBackoff wait.Backoff) (tokenReviewer, error) {
	panic("not implemented")
}

type tokenReviewV1Client struct {
	client rest.Interface
}

func (c *tokenReviewV1Client) Create(ctx context.Context, tokenReview *authenticationv1.TokenReview, opts metav1.CreateOptions) (result *authenticationv1.TokenReview, statusCode int, err error) {
	panic("not implemented")
}
