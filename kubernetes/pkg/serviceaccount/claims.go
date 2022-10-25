package serviceaccount

import (
	"context"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"

	apiserverserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/kubernetes/pkg/apis/core"
)

const (
	WarnOnlyBoundTokenExpirationSeconds = 60*60 + 7
	ExpirationExtensionSeconds          = 24 * 365 * 60 * 60
)

// time.Now stubbed out to allow testing
var now = time.Now

type privateClaims struct {
	Kubernetes kubernetes `json:"kubernetes.io,omitempty"`
}

type kubernetes struct {
	Namespace string          `json:"namespace,omitempty"`
	Svcacct   ref             `json:"serviceaccount,omitempty"`
	Pod       *ref            `json:"pod,omitempty"`
	Secret    *ref            `json:"secret,omitempty"`
	WarnAfter jwt.NumericDate `json:"warnafter,omitempty"`
}

type ref struct {
	Name string `json:"name,omitempty"`
	UID  string `json:"uid,omitempty"`
}

func Claims(sa core.ServiceAccount, pod *core.Pod, secret *core.Secret, expirationSeconds, warnafter int64, audience []string) (*jwt.Claims, interface{}) {
	panic("not implemented")
}

func NewValidator(getter ServiceAccountTokenGetter) Validator {
	return &validator{
		getter: getter,
	}
}

type validator struct {
	getter ServiceAccountTokenGetter
}

func (v *validator) Validate(ctx context.Context, _ string, public *jwt.Claims, privateObj interface{}) (*apiserverserviceaccount.ServiceAccountInfo, error) {
	panic("not implemented")
}

func (v *validator) NewPrivateClaims() interface{} {
	return &privateClaims{}
}
