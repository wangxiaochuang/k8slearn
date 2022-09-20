package yaml

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// DefaultMetaFactory is a default factory for versioning objects in JSON or
// YAML. The object in memory and in the default serialization will use the
// "kind" and "apiVersion" fields.
var DefaultMetaFactory = SimpleMetaFactory{}

// SimpleMetaFactory provides default methods for retrieving the type and version of objects
// that are identified with an "apiVersion" and "kind" fields in their JSON
// serialization. It may be parameterized with the names of the fields in memory, or an
// optional list of base structs to search for those fields in memory.
type SimpleMetaFactory struct{}

// Interpret will return the APIVersion and Kind of the JSON wire-format
// encoding of an object, or an error.
func (SimpleMetaFactory) Interpret(data []byte) (*schema.GroupVersionKind, error) {
	gvk := runtime.TypeMeta{}
	if err := yaml.Unmarshal(data, &gvk); err != nil {
		return nil, fmt.Errorf("could not interpret GroupVersionKind; unmarshal error: %v", err)
	}
	gv, err := schema.ParseGroupVersion(gvk.APIVersion)
	if err != nil {
		return nil, err
	}
	return &schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: gvk.Kind}, nil
}
