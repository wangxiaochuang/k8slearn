package authenticator

import (
	"context"
	"net/http"
)

func authenticate(ctx context.Context, implicitAuds Audiences, authenticate func() (*Response, bool, error)) (*Response, bool, error) {
	panic("not implemented")
}

type audAgnosticRequestAuthenticator struct {
	implicit Audiences
	delegate Request
}

var _ = Request(&audAgnosticRequestAuthenticator{})

func (a *audAgnosticRequestAuthenticator) AuthenticateRequest(req *http.Request) (*Response, bool, error) {
	return authenticate(req.Context(), a.implicit, func() (*Response, bool, error) {
		return a.delegate.AuthenticateRequest(req)
	})
}

func WrapAudienceAgnosticRequest(implicit Audiences, delegate Request) Request {
	return &audAgnosticRequestAuthenticator{
		implicit: implicit,
		delegate: delegate,
	}
}

type audAgnosticTokenAuthenticator struct {
	implicit Audiences
	delegate Token
}

var _ = Token(&audAgnosticTokenAuthenticator{})

func (a *audAgnosticTokenAuthenticator) AuthenticateToken(ctx context.Context, tok string) (*Response, bool, error) {
	return authenticate(ctx, a.implicit, func() (*Response, bool, error) {
		return a.delegate.AuthenticateToken(ctx, tok)
	})
}

func WrapAudienceAgnosticToken(implicit Audiences, delegate Token) Token {
	return &audAgnosticTokenAuthenticator{
		implicit: implicit,
		delegate: delegate,
	}
}
