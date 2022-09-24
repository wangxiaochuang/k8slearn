package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type SecuritySchemeProps struct {
	Description      string            `json:"description,omitempty"`
	Type             string            `json:"type"`
	Name             string            `json:"name,omitempty"`             // api key
	In               string            `json:"in,omitempty"`               // api key
	Flow             string            `json:"flow,omitempty"`             // oauth2
	AuthorizationURL string            `json:"authorizationUrl,omitempty"` // oauth2
	TokenURL         string            `json:"tokenUrl,omitempty"`         // oauth2
	Scopes           map[string]string `json:"scopes,omitempty"`           // oauth2
}

type SecurityScheme struct {
	VendorExtensible
	SecuritySchemeProps
}

func (s SecurityScheme) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.SecuritySchemeProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (s *SecurityScheme) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.SecuritySchemeProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &s.VendorExtensible)
}
