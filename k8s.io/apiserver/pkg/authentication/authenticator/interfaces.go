package authenticator

import (
	"context"
	"net/http"
)

type Token interface {
	AuthenticateToken(ctx context.Context, token string) (*Response, bool, error)
}

type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

type Response struct {
}
