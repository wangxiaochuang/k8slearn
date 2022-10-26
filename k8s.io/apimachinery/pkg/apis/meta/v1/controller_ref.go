package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func IsControlledBy(obj Object, owner Object) bool {
	ref := GetControllerOfNoCopy(obj)
	if ref == nil {
		return false
	}
	return ref.UID == owner.GetUID()
}

func GetControllerOf(controllee Object) *OwnerReference {
	ref := GetControllerOfNoCopy(controllee)
	if ref == nil {
		return nil
	}
	cp := *ref
	return &cp
}

func GetControllerOfNoCopy(controllee Object) *OwnerReference {
	refs := controllee.GetOwnerReferences()
	for i := range refs {
		if refs[i].Controller != nil && *refs[i].Controller {
			return &refs[i]
		}
	}
	return nil
}

func NewControllerRef(owner Object, gvk schema.GroupVersionKind) *OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return &OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
