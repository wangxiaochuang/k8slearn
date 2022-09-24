package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

const (
	jsonRef = "$ref"
)

type SimpleSchema struct {
	Type             string      `json:"type,omitempty"`
	Nullable         bool        `json:"nullable,omitempty"`
	Format           string      `json:"format,omitempty"`
	Items            *Items      `json:"items,omitempty"`
	CollectionFormat string      `json:"collectionFormat,omitempty"`
	Default          interface{} `json:"default,omitempty"`
	Example          interface{} `json:"example,omitempty"`
}

type CommonValidations struct {
	Maximum          *float64      `json:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int64        `json:"maxLength,omitempty"`
	MinLength        *int64        `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int64        `json:"maxItems,omitempty"`
	MinItems         *int64        `json:"minItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty"`
	MultipleOf       *float64      `json:"multipleOf,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
}

type Items struct {
	Refable
	CommonValidations
	SimpleSchema
	VendorExtensible
}

func (i *Items) UnmarshalJSON(data []byte) error {
	var validations CommonValidations
	if err := json.Unmarshal(data, &validations); err != nil {
		return err
	}
	var ref Refable
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}
	var simpleSchema SimpleSchema
	if err := json.Unmarshal(data, &simpleSchema); err != nil {
		return err
	}
	var vendorExtensible VendorExtensible
	if err := json.Unmarshal(data, &vendorExtensible); err != nil {
		return err
	}
	i.Refable = ref
	i.CommonValidations = validations
	i.SimpleSchema = simpleSchema
	i.VendorExtensible = vendorExtensible
	return nil
}

// MarshalJSON converts this items object to JSON
func (i Items) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(i.CommonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(i.SimpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(i.Refable)
	if err != nil {
		return nil, err
	}
	b4, err := json.Marshal(i.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b4, b3, b1, b2), nil
}
