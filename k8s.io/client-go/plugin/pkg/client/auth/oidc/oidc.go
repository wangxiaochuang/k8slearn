package oidc

import (
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func init() {
	if err := restclient.RegisterAuthProviderPlugin("oidc", newOIDCAuthProvider); err != nil {
		klog.Fatalf("Failed to register oidc auth plugin: %v", err)
	}
}

// p111
func newOIDCAuthProvider(clusterAddress string, cfg map[string]string, persister restclient.AuthProviderConfigPersister) (restclient.AuthProvider, error) {
	panic("not implemented")
}
