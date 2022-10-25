package union

import (
	"context"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

// unionAuthTokenHandler authenticates tokens using a chain of authenticator.Token objects
type unionAuthTokenHandler struct {
	// Handlers is a chain of request authenticators to delegate to
	Handlers []authenticator.Token
	// FailOnError determines whether an error returns short-circuits the chain
	FailOnError bool
}

func New(authTokenHandlers ...authenticator.Token) authenticator.Token {
	if len(authTokenHandlers) == 1 {
		return authTokenHandlers[0]
	}
	return &unionAuthTokenHandler{Handlers: authTokenHandlers, FailOnError: false}
}

func NewFailOnError(authTokenHandlers ...authenticator.Token) authenticator.Token {
	if len(authTokenHandlers) == 1 {
		return authTokenHandlers[0]
	}
	return &unionAuthTokenHandler{Handlers: authTokenHandlers, FailOnError: true}
}

func (authHandler *unionAuthTokenHandler) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
