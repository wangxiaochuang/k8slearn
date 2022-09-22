package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const GroupName = "meta.k8s.io"

var (
	schemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &schemeBuilder
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

var Unversioned = schema.GroupVersion{Group: "", Version: "v1"}

const WatchEventKind = "WatchEvent"

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

var scheme = runtime.NewScheme()

var ParameterCodec = runtime.NewParameterCodec(scheme)

var optionsTypes = []runtime.Object{
	&ListOptions{},
	&GetOptions{},
	&DeleteOptions{},
	&CreateOptions{},
	&UpdateOptions{},
	&PatchOptions{},
}

func AddToGroupVersion(scheme *runtime.Scheme, groupVersion schema.GroupVersion) {
	// 添加公共的Kind
	scheme.AddKnownTypeWithName(groupVersion.WithKind(WatchEventKind), &WatchEvent{})
	scheme.AddKnownTypeWithName(
		schema.GroupVersion{Group: groupVersion.Group, Version: runtime.APIVersionInternal}.WithKind(WatchEventKind),
		&InternalEvent{},
	)
	// 里面还是调用 AddKnownTypeWithName
	scheme.AddKnownTypes(groupVersion, optionsTypes...)
	// 没有group，只有version
	scheme.AddUnversionedTypes(Unversioned,
		&Status{},
		&APIVersions{},
		&APIGroupList{},
		&APIGroup{},
		&APIResourceList{},
	)
	// 非常基础的类型间的转换
	utilruntime.Must(RegisterConversions(scheme))
	utilruntime.Must(RegisterDefaults(scheme))
}
