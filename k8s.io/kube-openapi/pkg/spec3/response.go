package spec3

import (
	"encoding/json"
	"strconv"

	"github.com/go-openapi/swag"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Responses struct {
	ResponsesProps
	spec.VendorExtensible
}

func (r *Responses) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.ResponsesProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

func (r *Responses) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.ResponsesProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.VendorExtensible); err != nil {
		return err
	}

	return nil
}

type ResponsesProps struct {
	Default             *Response         `json:"-"`
	StatusCodeResponses map[int]*Response `json:"-"`
}

func (r ResponsesProps) MarshalJSON() ([]byte, error) {
	toser := map[string]*Response{}
	if r.Default != nil {
		toser["default"] = r.Default
	}
	for k, v := range r.StatusCodeResponses {
		toser[strconv.Itoa(k)] = v
	}
	return json.Marshal(toser)
}

func (r *ResponsesProps) UnmarshalJSON(data []byte) error {
	var res map[string]*Response
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}
	if v, ok := res["default"]; ok {
		r.Default = v
		delete(res, "default")
	}
	for k, v := range res {
		if nk, err := strconv.Atoi(k); err == nil {
			if r.StatusCodeResponses == nil {
				r.StatusCodeResponses = map[int]*Response{}
			}
			r.StatusCodeResponses[nk] = v
		}
	}
	return nil
}

type Response struct {
	spec.Refable
	ResponseProps
	spec.VendorExtensible
}

func (r *Response) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.ResponseProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

func (r *Response) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.ResponseProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.VendorExtensible); err != nil {
		return err
	}

	return nil
}

type ResponseProps struct {
	Description string                `json:"description,omitempty"`
	Headers     map[string]*Header    `json:"headers,omitempty"`
	Content     map[string]*MediaType `json:"content,omitempty"`
	Links       map[string]*Link      `json:"links,omitempty"`
}

type Link struct {
	spec.Refable
	LinkProps
	spec.VendorExtensible
}

func (r *Link) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.LinkProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

func (r *Link) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.LinkProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.VendorExtensible); err != nil {
		return err
	}

	return nil
}

type LinkProps struct {
	OperationId string                 `json:"operationId,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description,omitempty"`
	RequestBody interface{}            `json:"requestBody,omitempty"`
	Server      *Server                `json:"server,omitempty"`
}
