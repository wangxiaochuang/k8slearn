package openapi

import (
	"strings"

	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/schemamutation"
	"k8s.io/kube-openapi/pkg/validation/spec"

	genericfeatures "k8s.io/apiserver/pkg/features"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

const enumTypeDescriptionHeader = "Possible enum values:"

func GetOpenAPIDefinitionsWithoutDisabledFeatures(GetOpenAPIDefinitions common.GetOpenAPIDefinitions) common.GetOpenAPIDefinitions {
	return func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
		defs := GetOpenAPIDefinitions(ref)
		restoreDefinitions(defs)
		return defs
	}
}

func restoreDefinitions(defs map[string]common.OpenAPIDefinition) {
	// revert changes from OpenAPIEnums
	if !utilfeature.DefaultFeatureGate.Enabled(genericfeatures.OpenAPIEnums) {
		for gvk, def := range defs {
			orig := &def.Schema
			if ret := pruneEnums(orig); ret != orig {
				def.Schema = *ret
				defs[gvk] = def
			}
		}
	}
}

func pruneEnums(schema *spec.Schema) *spec.Schema {
	walker := schemamutation.Walker{
		SchemaCallback: func(schema *spec.Schema) *spec.Schema {
			orig := schema
			clone := func() {
				if orig == schema { // if schema has not been mutated yet
					schema = new(spec.Schema)
					*schema = *orig // make a clone from orig to schema
				}
			}
			if headerIndex := strings.Index(schema.Description, enumTypeDescriptionHeader); headerIndex != -1 {
				// remove the enum section from description.
				// note that the new lines before the header should be removed too,
				// thus the slice range.
				clone()
				schema.Description = schema.Description[:headerIndex]
			}
			if len(schema.Enum) != 0 {
				// remove the enum field
				clone()
				schema.Enum = nil
			}
			return schema
		},
		RefCallback: schemamutation.RefCallbackNoop,
	}
	return walker.WalkSchema(schema)
}
