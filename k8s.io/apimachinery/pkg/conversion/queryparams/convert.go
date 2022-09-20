package queryparams

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

type Marshaler interface {
	MarshalQueryParameter() (string, error)
}

type Unmarshaler interface {
	UnmarshalQueryParameter(string) error
}

func jsonTag(field reflect.StructField) (string, bool) {
	structTag := field.Tag.Get("json")
	if len(structTag) == 0 {
		return "", false
	}
	parts := strings.Split(structTag, ",")
	tag := parts[0]
	if tag == "-" {
		tag = ""
	}
	omitempty := false
	parts = parts[1:]
	for _, part := range parts {
		if part == "omitempty" {
			omitempty = true
			break
		}
	}
	return tag, omitempty
}

func isPointerKind(kind reflect.Kind) bool {
	return kind == reflect.Ptr
}

func isStructKind(kind reflect.Kind) bool {
	return kind == reflect.Struct
}

func isValueKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
		reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

func zeroValue(value reflect.Value) bool {
	return reflect.DeepEqual(reflect.Zero(value.Type()).Interface(), value.Interface())
}

func customMarshalValue(value reflect.Value) (reflect.Value, bool) {
	if !value.CanInterface() {
		return reflect.Value{}, false
	}

	marshaler, ok := value.Interface().(Marshaler)
	if !ok {
		if !isPointerKind(value.Kind()) && value.CanAddr() {
			marshaler, ok = value.Addr().Interface().(Marshaler)
			if !ok {
				return reflect.Value{}, false
			} else {
				return reflect.Value{}, false
			}
		}
	}

	if isPointerKind(value.Kind()) && zeroValue(value) {
		return reflect.ValueOf(""), true
	}

	v, err := marshaler.MarshalQueryParameter()
	if err != nil {
		return reflect.Value{}, false
	}
	return reflect.ValueOf(v), true
}

func addParam(values url.Values, tag string, omitempty bool, value reflect.Value) {
	if omitempty && zeroValue(value) {
		return
	}
	val := ""
	iValue := fmt.Sprintf("%v", value.Interface())
	if iValue != "<nil>" {
		val = iValue
	}
	values.Add(tag, val)
}

func addListOfParams(values url.Values, tag string, omitempty bool, list reflect.Value) {
	for i := 0; i < list.Len(); i++ {
		addParam(values, tag, omitempty, list.Index(i))
	}
}

func Convert(obj interface{}) (url.Values, error) {
	result := url.Values{}
	if obj == nil {
		return result, nil
	}
	var sv reflect.Value
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Ptr, reflect.Interface:
		sv = reflect.ValueOf(obj).Elem()
	default:
		return nil, fmt.Errorf("expecting a pointer or interface")
	}
	st := sv.Type()
	if !isStructKind(st.Kind()) {
		return nil, fmt.Errorf("expecting a pointer to a struct")
	}

	convertStruct(result, st, sv)

	return result, nil
}

func convertStruct(result url.Values, st reflect.Type, sv reflect.Value) {
	for i := 0; i < st.NumField(); i++ {
		field := sv.Field(i)
		tag, omitempty := jsonTag(st.Field(i))
		if len(tag) == 0 {
			continue
		}
		ft := field.Type()
		kind := ft.Kind()
		if isPointerKind(kind) {
			ft = ft.Elem()
			kind = ft.Kind()
			if !field.IsNil() {
				field = reflect.Indirect(field)
				omitempty = false
			}
		}

		switch {
		case isValueKind(kind):
			addParam(result, tag, omitempty, field)
		case kind == reflect.Array || kind == reflect.Slice:
			if isValueKind(ft.Elem().Kind()) {
				addListOfParams(result, tag, omitempty, field)
			}
		case isStructKind(kind) && !(zeroValue(field) && omitempty):
			if marshalValue, ok := customMarshalValue(field); ok {
				addParam(result, tag, omitempty, marshalValue)
			} else {
				convertStruct(result, ft, field)
			}
		}

	}
}
