package spec

import (
	"encoding/json"
	"strings"

	"github.com/go-openapi/swag"
)

type Paths struct {
	VendorExtensible
	Paths map[string]PathItem `json:"-"` // custom serializer to flatten this, each entry must start with "/"
}

func (p *Paths) UnmarshalJSON(data []byte) error {
	var res map[string]json.RawMessage
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	for k, v := range res {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			if p.Extensions == nil {
				p.Extensions = make(map[string]interface{})
			}
			var d interface{}
			if err := json.Unmarshal(v, &d); err != nil {
				return err
			}
			p.Extensions[k] = d
		}
		if strings.HasPrefix(k, "/") {
			if p.Paths == nil {
				p.Paths = make(map[string]PathItem)
			}
			var pi PathItem
			if err := json.Unmarshal(v, &pi); err != nil {
				return err
			}
			p.Paths[k] = pi
		}
	}
	return nil
}

// MarshalJSON converts this items object to JSON
func (p Paths) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(p.VendorExtensible)
	if err != nil {
		return nil, err
	}

	pths := make(map[string]PathItem)
	for k, v := range p.Paths {
		if strings.HasPrefix(k, "/") {
			pths[k] = v
		}
	}
	b2, err := json.Marshal(pths)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}
