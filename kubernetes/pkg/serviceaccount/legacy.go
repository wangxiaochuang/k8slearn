package serviceaccount

import (
	"context"

	"gopkg.in/square/go-jose.v2/jwt"

	v1 "k8s.io/api/core/v1"
	apiserverserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
)

func LegacyClaims(serviceAccount v1.ServiceAccount, secret v1.Secret) (*jwt.Claims, interface{}) {
	return &jwt.Claims{
			Subject: apiserverserviceaccount.MakeUsername(serviceAccount.Namespace, serviceAccount.Name),
		}, &legacyPrivateClaims{
			Namespace:          serviceAccount.Namespace,
			ServiceAccountName: serviceAccount.Name,
			ServiceAccountUID:  string(serviceAccount.UID),
			SecretName:         secret.Name,
		}
}

const LegacyIssuer = "kubernetes/serviceaccount"

type legacyPrivateClaims struct {
	ServiceAccountName string `json:"kubernetes.io/serviceaccount/service-account.name"`
	ServiceAccountUID  string `json:"kubernetes.io/serviceaccount/service-account.uid"`
	SecretName         string `json:"kubernetes.io/serviceaccount/secret.name"`
	Namespace          string `json:"kubernetes.io/serviceaccount/namespace"`
}

func NewLegacyValidator(lookup bool, getter ServiceAccountTokenGetter) Validator {
	return &legacyValidator{
		lookup: lookup,
		getter: getter,
	}
}

type legacyValidator struct {
	lookup bool
	getter ServiceAccountTokenGetter
}

var _ = Validator(&legacyValidator{})

func (v *legacyValidator) Validate(ctx context.Context, tokenData string, public *jwt.Claims, privateObj interface{}) (*apiserverserviceaccount.ServiceAccountInfo, error) {
	panic("not implemented")
}

func (v *legacyValidator) NewPrivateClaims() interface{} {
	return &legacyPrivateClaims{}
}
