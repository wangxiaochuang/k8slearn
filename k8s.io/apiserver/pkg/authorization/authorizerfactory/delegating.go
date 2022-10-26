package authorizerfactory

import (
	"errors"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/plugin/pkg/authorizer/webhook"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type DelegatingAuthorizerConfig struct {
	SubjectAccessReviewClient authorizationclient.AuthorizationV1Interface

	AllowCacheTTL time.Duration

	DenyCacheTTL time.Duration

	WebhookRetryBackoff *wait.Backoff
}

func (c DelegatingAuthorizerConfig) New() (authorizer.Authorizer, error) {
	if c.WebhookRetryBackoff == nil {
		return nil, errors.New("retry backoff parameters for delegating authorization webhook has not been specified")
	}

	return webhook.NewFromInterface(
		c.SubjectAccessReviewClient,
		c.AllowCacheTTL,
		c.DenyCacheTTL,
		*c.WebhookRetryBackoff,
		webhook.AuthorizerMetrics{
			RecordRequestTotal:   RecordRequestTotal,
			RecordRequestLatency: RecordRequestLatency,
		},
	)
}
