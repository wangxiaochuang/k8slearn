package authenticator

import (
	"context"
	"net/http"

	"k8s.io/apiserver/pkg/authentication/user"
)

type Token interface {
	AuthenticateToken(ctx context.Context, token string) (*Response, bool, error)
}

type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

type TokenFunc func(ctx context.Context, token string) (*Response, bool, error)

// AuthenticateToken implements authenticator.Token.
func (f TokenFunc) AuthenticateToken(ctx context.Context, token string) (*Response, bool, error) {
	return f(ctx, token)
}

// RequestFunc is a function that implements the Request interface.
type RequestFunc func(req *http.Request) (*Response, bool, error)

// AuthenticateRequest implements authenticator.Request.
func (f RequestFunc) AuthenticateRequest(req *http.Request) (*Response, bool, error) {
	return f(req)
}

type Response struct {
	Audiences Audiences
	User      user.Info
}
