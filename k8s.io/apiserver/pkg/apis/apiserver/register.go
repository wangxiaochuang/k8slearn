package apiserver

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const LegacyGroupName = "apiserver.k8s.io"
const GroupName = "apiserver.config.k8s.io"

var LegacySchemeGroupVersion = schema.GroupVersion{Group: LegacyGroupName, Version: runtime.APIVersionInternal}
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// 调用AddToScheme实际就是依次调用Builder注册的各个函数，会传递scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(LegacySchemeGroupVersion,
		&AdmissionConfiguration{},
		&EgressSelectorConfiguration{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&AdmissionConfiguration{},
		&EgressSelectorConfiguration{},
		&TracingConfiguration{},
	)
	return nil
}
