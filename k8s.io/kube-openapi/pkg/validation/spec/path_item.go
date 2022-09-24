package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type PathItemProps struct {
	Get        *Operation  `json:"get,omitempty"`
	Put        *Operation  `json:"put,omitempty"`
	Post       *Operation  `json:"post,omitempty"`
	Delete     *Operation  `json:"delete,omitempty"`
	Options    *Operation  `json:"options,omitempty"`
	Head       *Operation  `json:"head,omitempty"`
	Patch      *Operation  `json:"patch,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}

type PathItem struct {
	Refable
	VendorExtensible
	PathItemProps
}

func (p *PathItem) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.VendorExtensible); err != nil {
		return err
	}
	return json.Unmarshal(data, &p.PathItemProps)
}

func (p PathItem) MarshalJSON() ([]byte, error) {
	b3, err := json.Marshal(p.Refable)
	if err != nil {
		return nil, err
	}
	b4, err := json.Marshal(p.VendorExtensible)
	if err != nil {
		return nil, err
	}
	b5, err := json.Marshal(p.PathItemProps)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b3, b4, b5)
	return concated, nil
}
