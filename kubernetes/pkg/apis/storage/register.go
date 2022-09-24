package storage

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "storage.k8s.io"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&StorageClass{},
		&StorageClassList{},
		&VolumeAttachment{},
		&VolumeAttachmentList{},
		&CSINode{},
		&CSINodeList{},
		&CSIDriver{},
		&CSIDriverList{},
		&CSIStorageCapacity{},
		&CSIStorageCapacityList{},
	)
	return nil
}
