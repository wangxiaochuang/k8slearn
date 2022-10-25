package bearertoken

import (
	"errors"
	"net/http"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

type Authenticator struct {
	auth authenticator.Token
}

func New(auth authenticator.Token) *Authenticator {
	return &Authenticator{auth}
}

var invalidToken = errors.New("invalid bearer token")

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
