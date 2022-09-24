package install

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/apiserver/pkg/apis/apiserver/v1"
	"k8s.io/apiserver/pkg/apis/apiserver/v1alpha1"
	"k8s.io/apiserver/pkg/apis/apiserver/v1beta1"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/apis/storage/v1"
)

func init() {
	Install(legacyscheme.Scheme)
}

func Install(scheme *runtime.Scheme) {
	utilruntime.Must(storage.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	// utilruntime.Must(v1beta1.AddToScheme(scheme))
	// utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(scheme.SetVersionPriority(v1.SchemeGroupVersion, v1beta1.SchemeGroupVersion, v1alpha1.SchemeGroupVersion))
}
