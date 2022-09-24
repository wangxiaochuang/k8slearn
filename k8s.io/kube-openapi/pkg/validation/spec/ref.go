package spec

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-openapi/jsonreference"
)

type Refable struct {
	Ref Ref
}

func (r Refable) MarshalJSON() ([]byte, error) {
	return r.Ref.MarshalJSON()
}

func (r *Refable) UnmarshalJSON(d []byte) error {
	return json.Unmarshal(d, &r.Ref)
}

type Ref struct {
	jsonreference.Ref
}

func (r *Ref) RemoteURI() string {
	if r.String() == "" {
		return r.String()
	}

	u := *r.GetURL()
	u.Fragment = ""
	return u.String()
}

func (r *Ref) IsValidURI(basepaths ...string) bool {
	if r.String() == "" {
		return true
	}

	v := r.RemoteURI()
	if v == "" {
		return true
	}

	if r.HasFullURL {
		rr, err := http.Get(v)
		if err != nil {
			return false
		}

		return rr.StatusCode/100 == 2
	}

	if !(r.HasFileScheme || r.HasFullFilePath || r.HasURLPathOnly) {
		return false
	}

	// check for local file
	pth := v
	if r.HasURLPathOnly {
		base := "."
		if len(basepaths) > 0 {
			base = filepath.Dir(filepath.Join(basepaths...))
		}
		p, e := filepath.Abs(filepath.ToSlash(filepath.Join(base, pth)))
		if e != nil {
			return false
		}
		pth = p
	}

	fi, err := os.Stat(filepath.ToSlash(pth))
	if err != nil {
		return false
	}

	return !fi.IsDir()
}

func (r *Ref) Inherits(child Ref) (*Ref, error) {
	ref, err := r.Ref.Inherits(child.Ref)
	if err != nil {
		return nil, err
	}
	return &Ref{Ref: *ref}, nil
}

func NewRef(refURI string) (Ref, error) {
	ref, err := jsonreference.New(refURI)
	if err != nil {
		return Ref{}, err
	}
	return Ref{Ref: ref}, nil
}

func MustCreateRef(refURI string) Ref {
	return Ref{Ref: jsonreference.MustCreateRef(refURI)}
}

func (r Ref) MarshalJSON() ([]byte, error) {
	str := r.String()
	if str == "" {
		if r.IsRoot() {
			return []byte(`{"$ref":""}`), nil
		}
		return []byte("{}"), nil
	}
	v := map[string]interface{}{"$ref": str}
	return json.Marshal(v)
}

func (r *Ref) UnmarshalJSON(d []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(d, &v); err != nil {
		return err
	}
	return r.fromMap(v)
}

func (r *Ref) fromMap(v map[string]interface{}) error {
	if v == nil {
		return nil
	}

	if vv, ok := v["$ref"]; ok {
		if str, ok := vv.(string); ok {
			ref, err := jsonreference.New(str)
			if err != nil {
				return err
			}
			*r = Ref{Ref: ref}
		}
	}

	return nil
}
