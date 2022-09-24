package storage

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
)

type StorageClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Provisioner          string
	Parameters           map[string]string
	ReclaimPolicy        *api.PersistentVolumeReclaimPolicy
	MountOptions         []string
	AllowVolumeExpansion *bool
	VolumeBindingMode    *VolumeBindingMode
	AllowedTopologies    []api.TopologySelectorTerm
}

type StorageClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []StorageClass
}

type VolumeAttachment struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec   VolumeAttachmentSpec
	Status VolumeAttachmentStatus
}

type VolumeAttachmentList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []VolumeAttachment
}

type VolumeAttachmentSpec struct {
	Attacher string
	Source   VolumeAttachmentSource
	NodeName string
}

type VolumeAttachmentSource struct {
	PersistentVolumeName *string
	InlineVolumeSpec     *api.PersistentVolumeSpec
}
