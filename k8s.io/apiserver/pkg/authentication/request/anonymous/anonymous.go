package anonymous

import (
	"net/http"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

const (
	anonymousUser = user.Anonymous

	unauthenticatedGroup = user.AllUnauthenticated
)

func NewAuthenticator() authenticator.Request {
	return authenticator.RequestFunc(func(req *http.Request) (*authenticator.Response, bool, error) {
		auds, _ := authenticator.AudiencesFrom(req.Context())
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   anonymousUser,
				Groups: []string{unauthenticatedGroup},
			},
			Audiences: auds,
		}, true, nil
	})
}
