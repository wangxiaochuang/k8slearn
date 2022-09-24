package spec3

import (
	"encoding/json"

	"github.com/go-openapi/swag"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Encoding struct {
	EncodingProps
	spec.VendorExtensible
}

func (e *Encoding) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(e.EncodingProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(e.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.EncodingProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &e.VendorExtensible); err != nil {
		return err
	}
	return nil
}

type EncodingProps struct {
	ContentType   string             `json:"contentType,omitempty"`
	Headers       map[string]*Header `json:"headers,omitempty"`
	Style         string             `json:"style,omitempty"`
	Explode       string             `json:"explode,omitempty"`
	AllowReserved bool               `json:"allowReserved,omitempty"`
}
