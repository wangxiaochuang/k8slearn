package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

const (
	jsonArray = "array"
)

type HeaderProps struct {
	Description string `json:"description,omitempty"`
}

type Header struct {
	CommonValidations
	SimpleSchema
	VendorExtensible
	HeaderProps
}

func (h Header) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(h.CommonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(h.SimpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(h.HeaderProps)
	if err != nil {
		return nil, err
	}
	b4, err := json.Marshal(h.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3, b4), nil
}

func (h *Header) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &h.CommonValidations); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &h.SimpleSchema); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &h.VendorExtensible); err != nil {
		return err
	}
	return json.Unmarshal(data, &h.HeaderProps)
}
