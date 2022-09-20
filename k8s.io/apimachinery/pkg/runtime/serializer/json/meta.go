package json

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MetaFactory interface {
	Interpret(data []byte) (*schema.GroupVersionKind, error)
}

var DefaultMetaFactory = SimpleMetaFactory{}

type SimpleMetaFactory struct {
}

func (SimpleMetaFactory) Interpret(data []byte) (*schema.GroupVersionKind, error) {
	findKind := struct {
		// +optional
		APIVersion string `json:"apiVersion,omitempty"`
		// +optional
		Kind string `json:"kind,omitempty"`
	}{}
	if err := json.Unmarshal(data, &findKind); err != nil {
		return nil, fmt.Errorf("couldn't get version/kind; json parse error: %v", err)
	}
	gv, err := schema.ParseGroupVersion(findKind.APIVersion)
	if err != nil {
		return nil, err
	}
	return &schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: findKind.Kind}, nil
}
