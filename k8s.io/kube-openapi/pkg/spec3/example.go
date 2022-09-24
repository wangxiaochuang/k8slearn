package spec3

import (
	"encoding/json"

	"github.com/go-openapi/swag"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Example struct {
	spec.Refable
	ExampleProps
	spec.VendorExtensible
}

func (e *Example) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(e.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(e.ExampleProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(e.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

func (e *Example) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &e.ExampleProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &e.VendorExtensible); err != nil {
		return err
	}
	return nil
}

type ExampleProps struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}
