package serviceaccount

import (
	"context"

	"gopkg.in/square/go-jose.v2/jwt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	apiserverserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
)

type ServiceAccountTokenGetter interface {
	GetServiceAccount(namespace, name string) (*v1.ServiceAccount, error)
	GetPod(namespace, name string) (*v1.Pod, error)
	GetSecret(namespace, name string) (*v1.Secret, error)
}

type TokenGenerator interface {
	GenerateToken(claims *jwt.Claims, privateClaims interface{}) (string, error)
}

// p227
func JWTTokenAuthenticator(issuers []string, keys []interface{}, implicitAuds authenticator.Audiences, validator Validator) authenticator.Token {
	issuersMap := make(map[string]bool)
	for _, issuer := range issuers {
		issuersMap[issuer] = true
	}
	return &jwtTokenAuthenticator{
		issuers:      issuersMap,
		keys:         keys,
		implicitAuds: implicitAuds,
		validator:    validator,
	}
}

type jwtTokenAuthenticator struct {
	issuers      map[string]bool
	keys         []interface{}
	validator    Validator
	implicitAuds authenticator.Audiences
}

type Validator interface {
	Validate(ctx context.Context, tokenData string, public *jwt.Claims, private interface{}) (*apiserverserviceaccount.ServiceAccountInfo, error)
	NewPrivateClaims() interface{}
}

func (j *jwtTokenAuthenticator) AuthenticateToken(ctx context.Context, tokenData string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
