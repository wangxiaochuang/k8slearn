package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type TagProps struct {
	Description  string                 `json:"description,omitempty"`
	Name         string                 `json:"name,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

type Tag struct {
	VendorExtensible
	TagProps
}

func (t Tag) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(t.TagProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(t.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (t *Tag) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &t.TagProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &t.VendorExtensible)
}
