package internalversion

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type ListOptions struct {
	metav1.TypeMeta

	LabelSelector        labels.Selector
	FieldSelector        fields.Selector
	Watch                bool
	AllowWatchBookmarks  bool
	ResourceVersion      string
	ResourceVersionMatch metav1.ResourceVersionMatch
	TimeoutSeconds       *int64
	Limit                int64
	Continue             string
}

type List struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	Items []runtime.Object
}
