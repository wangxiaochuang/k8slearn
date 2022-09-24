package spec

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type OperationProps struct {
	Description  string                 `json:"description,omitempty"`
	Consumes     []string               `json:"consumes,omitempty"`
	Produces     []string               `json:"produces,omitempty"`
	Schemes      []string               `json:"schemes,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	ID           string                 `json:"operationId,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty"`
	Security     []map[string][]string  `json:"security,omitempty"`
	Parameters   []Parameter            `json:"parameters,omitempty"`
	Responses    *Responses             `json:"responses,omitempty"`
}

func (op OperationProps) MarshalJSON() ([]byte, error) {
	type Alias OperationProps
	if op.Security == nil {
		return json.Marshal(&struct {
			Security []map[string][]string `json:"security,omitempty"`
			*Alias
		}{
			Security: op.Security,
			Alias:    (*Alias)(&op),
		})
	}
	return json.Marshal(&struct {
		Security []map[string][]string `json:"security"`
		*Alias
	}{
		Security: op.Security,
		Alias:    (*Alias)(&op),
	})
}

type Operation struct {
	VendorExtensible
	OperationProps
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &o.OperationProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &o.VendorExtensible)
}

func (o Operation) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(o.OperationProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(o.VendorExtensible)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}
