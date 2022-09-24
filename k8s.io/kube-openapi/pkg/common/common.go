package common

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

const (
	ExtensionPrefix   = "x-kubernetes-"
	ExtensionV2Schema = ExtensionPrefix + "v2-schema"
)

type OpenAPIDefinition struct {
	Schema       spec.Schema
	Dependencies []string
}

type ReferenceCallback func(path string) spec.Ref

type GetOpenAPIDefinitions func(ReferenceCallback) map[string]OpenAPIDefinition

type OpenAPIDefinitionGetter interface {
	OpenAPIDefinition() *OpenAPIDefinition
}

type OpenAPIV3DefinitionGetter interface {
	OpenAPIV3Definition() *OpenAPIDefinition
}

type PathHandler interface {
	Handle(path string, handler http.Handler)
}

type PathHandlerByGroupVersion interface {
	Handle(path string, handler http.Handler)
	HandlePrefix(path string, handler http.Handler)
}

type Config struct {
	ProtocolList                   []string
	Info                           *spec.Info
	DefaultResponse                *spec.Response
	ResponseDefinitions            map[string]spec.Response
	CommonResponses                map[int]spec.Response
	IgnorePrefixes                 []string
	GetDefinitions                 GetOpenAPIDefinitions
	Definitions                    map[string]OpenAPIDefinition
	GetOperationIDAndTags          func(r *restful.Route) (string, []string, error)
	GetOperationIDAndTagsFromRoute func(r Route) (string, []string, error)
	GetDefinitionName              func(name string) (string, spec.Extensions)
	PostProcessSpec                func(*spec.Swagger) (*spec.Swagger, error)
	SecurityDefinitions            *spec.SecurityDefinitions
	DefaultSecurity                []map[string][]string
}

type OpenAPIV3Config struct {
	Info                *spec.Info
	DefaultResponse     *spec3.Response
	ResponseDefinitions map[string]*spec3.Response
}

// p212
type typeInfo struct {
	name   string
	format string
	zero   interface{}
}

var schemaTypeFormatMap = map[string]typeInfo{
	"uint":        {"integer", "int32", 0.},
	"uint8":       {"integer", "byte", 0.},
	"uint16":      {"integer", "int32", 0.},
	"uint32":      {"integer", "int64", 0.},
	"uint64":      {"integer", "int64", 0.},
	"int":         {"integer", "int32", 0.},
	"int8":        {"integer", "byte", 0.},
	"int16":       {"integer", "int32", 0.},
	"int32":       {"integer", "int32", 0.},
	"int64":       {"integer", "int64", 0.},
	"byte":        {"integer", "byte", 0},
	"float64":     {"number", "double", 0.},
	"float32":     {"number", "float", 0.},
	"bool":        {"boolean", "", false},
	"time.Time":   {"string", "date-time", ""},
	"string":      {"string", "", ""},
	"integer":     {"integer", "", 0.},
	"number":      {"number", "", 0.},
	"boolean":     {"boolean", "", false},
	"[]byte":      {"string", "byte", ""}, // base64 encoded characters
	"interface{}": {"object", "", interface{}(nil)},
}

func OpenAPITypeFormat(typeName string) (string, string) {
	mapped, ok := schemaTypeFormatMap[typeName]
	if !ok {
		return "", ""
	}
	return mapped.name, mapped.format
}

func OpenAPIZeroValue(typeName string) (interface{}, bool) {
	mapped, ok := schemaTypeFormatMap[typeName]
	if !ok {
		return nil, false
	}
	return mapped.zero, true
}

// p299
func EscapeJsonPointer(p string) string {
	// Escaping reference name using rfc6901
	p = strings.Replace(p, "~", "~0", -1)
	p = strings.Replace(p, "/", "~1", -1)
	return p
}

func EmbedOpenAPIDefinitionIntoV2Extension(main OpenAPIDefinition, embedded OpenAPIDefinition) OpenAPIDefinition {
	if main.Schema.Extensions == nil {
		main.Schema.Extensions = make(map[string]interface{})
	}
	main.Schema.Extensions[ExtensionV2Schema] = embedded.Schema
	return main
}

// GenerateOpenAPIV3OneOfSchema generate the set of schemas that MUST be assigned to SchemaProps.OneOf
func GenerateOpenAPIV3OneOfSchema(types []string) (oneOf []spec.Schema) {
	for _, t := range types {
		oneOf = append(oneOf, spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{t}}})
	}
	return
}
