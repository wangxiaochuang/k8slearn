package spec3

import (
	"encoding/json"

	"github.com/go-openapi/swag"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type MediaType struct {
	MediaTypeProps
	spec.VendorExtensible
}

func (m *MediaType) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(m.MediaTypeProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(m.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (m *MediaType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &m.MediaTypeProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m.VendorExtensible); err != nil {
		return err
	}
	return nil
}

type MediaTypeProps struct {
	Schema   *spec.Schema         `json:"schema,omitempty"`
	Example  interface{}          `json:"example,omitempty"`
	Examples map[string]*Example  `json:"examples,omitempty"`
	Encoding map[string]*Encoding `json:"encoding,omitempty"`
}
