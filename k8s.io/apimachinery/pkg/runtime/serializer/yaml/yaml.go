package yaml

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// yamlSerializer converts YAML passed to the Decoder methods to JSON.
type yamlSerializer struct {
	// the nested serializer
	runtime.Serializer
}

// yamlSerializer implements Serializer
var _ runtime.Serializer = yamlSerializer{}

// NewDecodingSerializer adds YAML decoding support to a serializer that supports JSON.
func NewDecodingSerializer(jsonSerializer runtime.Serializer) runtime.Serializer {
	return &yamlSerializer{jsonSerializer}
}

func (c yamlSerializer) Decode(data []byte, gvk *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	out, err := yaml.ToJSON(data)
	if err != nil {
		return nil, nil, err
	}
	data = out
	return c.Serializer.Decode(data, gvk, into)
}
