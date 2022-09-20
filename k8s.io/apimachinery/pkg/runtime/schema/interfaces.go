package schema

type ObjectKind interface {
	SetGroupVersionKind(kind GroupVersionKind)
	GroupVersionKind() GroupVersionKind
}

var EmptyObjectKind = emptyObjectKind{}

type emptyObjectKind struct{}

func (emptyObjectKind) SetGroupVersionKind(gvk GroupVersionKind) {}

func (emptyObjectKind) GroupVersionKind() GroupVersionKind { return GroupVersionKind{} }
