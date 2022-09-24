package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type ResponseProps struct {
	Description string                 `json:"description,omitempty"`
	Schema      *Schema                `json:"schema,omitempty"`
	Headers     map[string]Header      `json:"headers,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

type Response struct {
	Refable
	ResponseProps
	VendorExtensible
}

func (r *Response) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.ResponseProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.Refable); err != nil {
		return err
	}
	return json.Unmarshal(data, &r.VendorExtensible)
}

func (r Response) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.ResponseProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.Refable)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

func NewResponse() *Response {
	return new(Response)
}

func ResponseRef(url string) *Response {
	resp := NewResponse()
	resp.Ref = MustCreateRef(url)
	return resp
}
