package authorizer

import (
	"context"

	"k8s.io/apiserver/pkg/authentication/user"
)

type Attributes interface {
	GetUser() user.Info

	GetVerb() string

	IsReadOnly() bool

	GetNamespace() string

	GetResource() string

	GetSubresource() string

	GetName() string

	GetAPIGroup() string

	GetAPIVersion() string

	IsResourceRequest() bool

	GetPath() string
}

type Authorizer interface {
	Authorize(ctx context.Context, a Attributes) (authorized Decision, reason string, err error)
}

// p81
type RuleResolver interface {
	RulesFor(user user.Info, namespace string) ([]ResourceRuleInfo, []NonResourceRuleInfo, bool, error)
}

type NonResourceRuleInfo interface {
	GetVerbs() []string
	GetNonResourceURLs() []string
}

type Decision int
