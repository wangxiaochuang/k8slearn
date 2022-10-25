package group

import (
	"net/http"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

type AuthenticatedGroupAdder struct {
	// Authenticator is delegated to make the authentication decision
	Authenticator authenticator.Request
}

func NewAuthenticatedGroupAdder(auth authenticator.Request) authenticator.Request {
	return &AuthenticatedGroupAdder{auth}
}

func (g *AuthenticatedGroupAdder) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
