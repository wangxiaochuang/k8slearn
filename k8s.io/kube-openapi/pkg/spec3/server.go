package spec3

import (
	"encoding/json"

	"github.com/go-openapi/swag"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Server struct {
	ServerProps
	spec.VendorExtensible
}

type ServerProps struct {
	Description string                     `json:"description,omitempty"`
	URL         string                     `json:"url"`
	Variables   map[string]*ServerVariable `json:"variables,omitempty"`
}

func (s *Server) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.ServerProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (s *Server) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.ServerProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &s.VendorExtensible); err != nil {
		return err
	}
	return nil
}

type ServerVariable struct {
	ServerVariableProps
	spec.VendorExtensible
}

type ServerVariableProps struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

func (s *ServerVariable) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.ServerVariableProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (s *ServerVariable) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.ServerVariableProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &s.VendorExtensible); err != nil {
		return err
	}
	return nil
}
