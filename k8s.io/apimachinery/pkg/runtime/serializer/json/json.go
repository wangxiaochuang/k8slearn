package json

import (
	"encoding/json"
	"io"
	"strconv"

	kjson "sigs.k8s.io/json"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/recognizer"
	"k8s.io/apimachinery/pkg/util/framer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

func NewSerializer(meta MetaFactory, creater runtime.ObjectCreater, typer runtime.ObjectTyper, pretty bool) *Serializer {
	return NewSerializerWithOptions(meta, creater, typer, SerializerOptions{false, pretty, false})
}

func NewYAMLSerializer(meta MetaFactory, creater runtime.ObjectCreater, typer runtime.ObjectTyper) *Serializer {
	return NewSerializerWithOptions(meta, creater, typer, SerializerOptions{true, false, false})
}

func NewSerializerWithOptions(meta MetaFactory, creater runtime.ObjectCreater, typer runtime.ObjectTyper, options SerializerOptions) *Serializer {
	return &Serializer{
		meta:       meta,
		creater:    creater,
		typer:      typer,
		options:    options,
		identifier: identifier(options),
	}
}

func identifier(options SerializerOptions) runtime.Identifier {
	result := map[string]string{
		"name":   "json",
		"yaml":   strconv.FormatBool(options.Yaml),
		"pretty": strconv.FormatBool(options.Pretty),
		"strict": strconv.FormatBool(options.Strict),
	}
	identifier, err := json.Marshal(result)
	if err != nil {
		klog.Fatalf("Failed marshaling identifier for json Serializer: %v", err)
	}
	return runtime.Identifier(identifier)
}

type SerializerOptions struct {
	Yaml   bool
	Pretty bool
	Strict bool
}

type Serializer struct {
	meta    MetaFactory
	options SerializerOptions
	creater runtime.ObjectCreater
	typer   runtime.ObjectTyper

	identifier runtime.Identifier
}

var _ runtime.Serializer = &Serializer{}
var _ recognizer.RecognizingDecoder = &Serializer{}

func gvkWithDefaults(actual, defaultGVK schema.GroupVersionKind) schema.GroupVersionKind {
	if len(actual.Kind) == 0 {
		actual.Kind = defaultGVK.Kind
	}
	if len(actual.Version) == 0 && len(actual.Group) == 0 {
		actual.Group = defaultGVK.Group
		actual.Version = defaultGVK.Version
	}
	if len(actual.Version) == 0 && actual.Group == defaultGVK.Group {
		actual.Version = defaultGVK.Version
	}
	return actual
}

func (s *Serializer) Decode(originalData []byte, gvk *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	data := originalData
	if s.options.Yaml {
		altered, err := yaml.YAMLToJSON(data)
		if err != nil {
			return nil, nil, err
		}
		data = altered
	}

	actual, err := s.meta.Interpret(data)
	if err != nil {
		return nil, nil, err
	}

	if gvk != nil {
		*actual = gvkWithDefaults(*actual, *gvk)
	}

	if unk, ok := into.(*runtime.Unknown); ok && unk != nil {
		unk.Raw = originalData
		unk.ContentType = runtime.ContentTypeJSON
		unk.GetObjectKind().SetGroupVersionKind(*actual)
		return unk, actual, nil
	}

	if into != nil {
		_, isUnstructured := into.(runtime.Unstructured)
		types, _, err := s.typer.ObjectKinds(into)
		switch {
		case runtime.IsNotRegisteredError(err), isUnstructured:
			strictErrs, err := s.unmarshal(into, data, originalData)
			if err != nil {
				return nil, actual, err
			}

			if isUnstructured {
				*actual = into.GetObjectKind().GroupVersionKind()
				if len(actual.Kind) == 0 {
					return nil, actual, runtime.NewMissingKindErr(string(originalData))
				}
				// TODO(109023): require apiVersion here as well once unstructuredJSONScheme#Decode does
			}

			if len(strictErrs) > 0 {
				return into, actual, runtime.NewStrictDecodingError(strictErrs)
			}
			return into, actual, nil
		case err != nil:
			return nil, actual, err
		default:
			*actual = gvkWithDefaults(*actual, types[0])
		}
	}

	if len(actual.Kind) == 0 {
		return nil, actual, runtime.NewMissingKindErr(string(originalData))
	}
	if len(actual.Version) == 0 {
		return nil, actual, runtime.NewMissingVersionErr(string(originalData))
	}

	obj, err := runtime.UseOrCreateObject(s.typer, s.creater, *actual, into)
	if err != nil {
		return nil, actual, err
	}

	strictErrs, err := s.unmarshal(obj, data, originalData)
	if err != nil {
		return nil, actual, err
	} else if len(strictErrs) > 0 {
		return obj, actual, runtime.NewStrictDecodingError(strictErrs)
	}
	return obj, actual, nil
}

func (s *Serializer) Encode(obj runtime.Object, w io.Writer) error {
	if co, ok := obj.(runtime.CacheableObject); ok {
		return co.CacheEncode(s.Identifier(), s.doEncode, w)
	}
	return s.doEncode(obj, w)
}

func (s *Serializer) doEncode(obj runtime.Object, w io.Writer) error {
	if s.options.Yaml {
		json, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		data, err := yaml.JSONToYAML(json)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}

	if s.options.Pretty {
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	encoder := json.NewEncoder(w)
	return encoder.Encode(obj)
}

func (s *Serializer) IsStrict() bool {
	return s.options.Strict
}

func (s *Serializer) unmarshal(into runtime.Object, data, originalData []byte) (strictErrs []error, err error) {
	// If the deserializer is non-strict, return here.
	if !s.options.Strict {
		if err := kjson.UnmarshalCaseSensitivePreserveInts(data, into); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if s.options.Yaml {
		// In strict mode pass the original data through the YAMLToJSONStrict converter.
		// This is done to catch duplicate fields in YAML that would have been dropped in the original YAMLToJSON conversion.
		// TODO: rework YAMLToJSONStrict to return warnings about duplicate fields without terminating so we don't have to do this twice.
		_, err := yaml.YAMLToJSONStrict(originalData)
		if err != nil {
			strictErrs = append(strictErrs, err)
		}
	}

	var strictJSONErrs []error
	if u, isUnstructured := into.(runtime.Unstructured); isUnstructured {
		// Unstructured is a custom unmarshaler that gets delegated
		// to, so in order to detect strict JSON errors we need
		// to unmarshal directly into the object.
		m := map[string]interface{}{}
		strictJSONErrs, err = kjson.UnmarshalStrict(data, &m)
		u.SetUnstructuredContent(m)
	} else {
		strictJSONErrs, err = kjson.UnmarshalStrict(data, into)
	}
	if err != nil {
		// fatal decoding error, not due to strictness
		return nil, err
	}
	strictErrs = append(strictErrs, strictJSONErrs...)
	return strictErrs, nil
}

func (s *Serializer) Identifier() runtime.Identifier {
	return s.identifier
}

func (s *Serializer) RecognizesData(data []byte) (ok, unknown bool, err error) {
	if s.options.Yaml {
		// we could potentially look for '---'
		return false, true, nil
	}
	return utilyaml.IsJSONBuffer(data), false, nil
}

var Framer = jsonFramer{}

type jsonFramer struct{}

// NewFrameWriter implements stream framing for this serializer
func (jsonFramer) NewFrameWriter(w io.Writer) io.Writer {
	// we can write JSON objects directly to the writer, because they are self-framing
	return w
}

// NewFrameReader implements stream framing for this serializer
func (jsonFramer) NewFrameReader(r io.ReadCloser) io.ReadCloser {
	// we need to extract the JSON chunks of data to pass to Decode()
	return framer.NewJSONFramedReader(r)
}

var YAMLFramer = yamlFramer{}

type yamlFramer struct{}

// NewFrameWriter implements stream framing for this serializer
func (yamlFramer) NewFrameWriter(w io.Writer) io.Writer {
	return yamlFrameWriter{w}
}

// NewFrameReader implements stream framing for this serializer
func (yamlFramer) NewFrameReader(r io.ReadCloser) io.ReadCloser {
	// extract the YAML document chunks directly
	return utilyaml.NewDocumentDecoder(r)
}

type yamlFrameWriter struct {
	w io.Writer
}

func (w yamlFrameWriter) Write(data []byte) (n int, err error) {
	if _, err := w.w.Write([]byte("---\n")); err != nil {
		return 0, err
	}
	return w.w.Write(data)
}
