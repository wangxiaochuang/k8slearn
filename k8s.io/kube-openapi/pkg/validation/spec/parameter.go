package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type ParamProps struct {
	Description     string  `json:"description,omitempty"`
	Name            string  `json:"name,omitempty"`
	In              string  `json:"in,omitempty"`
	Required        bool    `json:"required,omitempty"`
	Schema          *Schema `json:"schema,omitempty"`
	AllowEmptyValue bool    `json:"allowEmptyValue,omitempty"`
}

type Parameter struct {
	Refable
	CommonValidations
	SimpleSchema
	VendorExtensible
	ParamProps
}

func (p *Parameter) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.CommonValidations); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.SimpleSchema); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.VendorExtensible); err != nil {
		return err
	}
	return json.Unmarshal(data, &p.ParamProps)
}

func (p Parameter) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(p.CommonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(p.SimpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(p.Refable)
	if err != nil {
		return nil, err
	}
	b4, err := json.Marshal(p.VendorExtensible)
	if err != nil {
		return nil, err
	}
	b5, err := json.Marshal(p.ParamProps)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b3, b1, b2, b4, b5), nil
}
