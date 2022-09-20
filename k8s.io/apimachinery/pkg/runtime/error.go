package runtime

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type notRegisteredErr struct {
	schemeName string
	gvk        schema.GroupVersionKind
	target     GroupVersioner
	t          reflect.Type
}

func NewNotRegisteredErrForKind(schemeName string, gvk schema.GroupVersionKind) error {
	return &notRegisteredErr{schemeName: schemeName, gvk: gvk}
}

func NewNotRegisteredErrForType(schemeName string, t reflect.Type) error {
	return &notRegisteredErr{schemeName: schemeName, t: t}
}

func NewNotRegisteredErrForTarget(schemeName string, t reflect.Type, target GroupVersioner) error {
	return &notRegisteredErr{schemeName: schemeName, t: t, target: target}
}

func NewNotRegisteredGVKErrForTarget(schemeName string, gvk schema.GroupVersionKind, target GroupVersioner) error {
	return &notRegisteredErr{schemeName: schemeName, gvk: gvk, target: target}
}

func (k *notRegisteredErr) Error() string {
	if k.t != nil && k.target != nil {
		return fmt.Sprintf("%v is not suitable for converting to %q in scheme %q", k.t, k.target, k.schemeName)
	}
	nullGVK := schema.GroupVersionKind{}
	if k.gvk != nullGVK && k.target != nil {
		return fmt.Sprintf("%q is not suitable for converting to %q in scheme %q", k.gvk.GroupVersion(), k.target, k.schemeName)
	}
	if k.t != nil {
		return fmt.Sprintf("no kind is registered for the type %v in scheme %q", k.t, k.schemeName)
	}
	if len(k.gvk.Kind) == 0 {
		return fmt.Sprintf("no version %q has been registered in scheme %q", k.gvk.GroupVersion(), k.schemeName)
	}
	if k.gvk.Version == APIVersionInternal {
		return fmt.Sprintf("no kind %q is registered for the internal version of group %q in scheme %q", k.gvk.Kind, k.gvk.Group, k.schemeName)
	}

	return fmt.Sprintf("no kind %q is registered for version %q in scheme %q", k.gvk.Kind, k.gvk.GroupVersion(), k.schemeName)
}

func IsNotRegisteredError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*notRegisteredErr)
	return ok
}

type missingKindErr struct {
	data string
}

func NewMissingKindErr(data string) error {
	return &missingKindErr{data}
}

func (k *missingKindErr) Error() string {
	return fmt.Sprintf("Object 'Kind' is missing in '%s'", k.data)
}

func IsMissingKind(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*missingKindErr)
	return ok
}

type missingVersionErr struct {
	data string
}

func NewMissingVersionErr(data string) error {
	return &missingVersionErr{data}
}

func (k *missingVersionErr) Error() string {
	return fmt.Sprintf("Object 'apiVersion' is missing in '%s'", k.data)
}

func IsMissingVersion(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*missingVersionErr)
	return ok
}

type strictDecodingError struct {
	errors []error
}

// NewStrictDecodingError creates a new strictDecodingError object.
func NewStrictDecodingError(errors []error) error {
	return &strictDecodingError{
		errors: errors,
	}
}

func (e *strictDecodingError) Error() string {
	var s strings.Builder
	s.WriteString("strict decoding error: ")
	for i, err := range e.errors {
		if i != 0 {
			s.WriteString(", ")
		}
		s.WriteString(err.Error())
	}
	return s.String()
}

func (e *strictDecodingError) Errors() []error {
	return e.errors
}

func IsStrictDecodingError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*strictDecodingError)
	return ok
}

func AsStrictDecodingError(err error) (*strictDecodingError, bool) {
	if err == nil {
		return nil, false
	}
	strictErr, ok := err.(*strictDecodingError)
	return strictErr, ok
}
