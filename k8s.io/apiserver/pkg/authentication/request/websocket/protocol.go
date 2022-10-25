package websocket

import (
	"errors"
	"net/http"
	"net/textproto"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

const bearerProtocolPrefix = "base64url.bearer.authorization.k8s.io."

var protocolHeader = textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Protocol")

var errInvalidToken = errors.New("invalid bearer token")

type ProtocolAuthenticator struct {
	// auth is the token authenticator to use to validate the token
	auth authenticator.Token
}

func NewProtocolAuthenticator(auth authenticator.Token) *ProtocolAuthenticator {
	return &ProtocolAuthenticator{auth}
}

func (a *ProtocolAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
