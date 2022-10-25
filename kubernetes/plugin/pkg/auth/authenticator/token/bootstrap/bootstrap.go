package bootstrap

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

func NewTokenAuthenticator(lister corev1listers.SecretNamespaceLister) *TokenAuthenticator {
	return &TokenAuthenticator{lister}
}

type TokenAuthenticator struct {
	lister corev1listers.SecretNamespaceLister
}

func tokenErrorf(s *corev1.Secret, format string, i ...interface{}) {
	format = fmt.Sprintf("Bootstrap secret %s/%s matching bearer token ", s.Namespace, s.Name) + format
	klog.V(3).Infof(format, i...)
}

func (t *TokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
