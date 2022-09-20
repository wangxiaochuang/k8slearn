package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)
