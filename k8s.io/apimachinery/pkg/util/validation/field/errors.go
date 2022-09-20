package field

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Error struct {
	Type     ErrorType
	Field    string
	BadValue interface{}
	Detail   string
}

var _ error = &Error{}

func (v *Error) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.ErrorBody())
}

type omitValueType struct{}

var omitValue = omitValueType{}

func (v *Error) ErrorBody() string {
	var s string
	switch {
	case v.Type == ErrorTypeRequired:
		s = v.Type.String()
	case v.Type == ErrorTypeForbidden:
		s = v.Type.String()
	case v.Type == ErrorTypeTooLong:
		s = v.Type.String()
	case v.Type == ErrorTypeInternal:
		s = v.Type.String()
	case v.BadValue == omitValue:
		s = v.Type.String()
	default:
		value := v.BadValue
		valueType := reflect.TypeOf(value)
		if value == nil || valueType == nil {
			value = "null"
		} else if valueType.Kind() == reflect.Ptr {
			if reflectValue := reflect.ValueOf(value); reflectValue.IsNil() {
				value = "null"
			} else {
				value = reflectValue.Elem().Interface()
			}
		}
		switch t := value.(type) {
		case int64, int32, float64, float32, bool:
			s = fmt.Sprintf("%s: %v", v.Type, value)
		case string:
			s = fmt.Sprintf("%s: %q", v.Type, t)
		case fmt.Stringer:
			s = fmt.Sprintf("%s: %s", v.Type, t.String())
		default:
			s = fmt.Sprintf("%s: %#v", v.Type, value)
		}
	}
	if len(v.Detail) != 0 {
		s += fmt.Sprintf(": %s", v.Detail)
	}
	return s
}

type ErrorType string

const (
	ErrorTypeNotFound     ErrorType = "FieldValueNotFound"
	ErrorTypeRequired     ErrorType = "FieldValueRequired"
	ErrorTypeDuplicate    ErrorType = "FieldValueDuplicate"
	ErrorTypeInvalid      ErrorType = "FieldValueInvalid"
	ErrorTypeNotSupported ErrorType = "FieldValueNotSupported"
	ErrorTypeForbidden    ErrorType = "FieldValueForbidden"
	ErrorTypeTooLong      ErrorType = "FieldValueTooLong"
	ErrorTypeTooMany      ErrorType = "FieldValueTooMany"
	ErrorTypeInternal     ErrorType = "InternalError"
	ErrorTypeTypeInvalid  ErrorType = "FieldValueTypeInvalid"
)

func (t ErrorType) String() string {
	switch t {
	case ErrorTypeNotFound:
		return "Not found"
	case ErrorTypeRequired:
		return "Required value"
	case ErrorTypeDuplicate:
		return "Duplicate value"
	case ErrorTypeInvalid:
		return "Invalid value"
	case ErrorTypeNotSupported:
		return "Unsupported value"
	case ErrorTypeForbidden:
		return "Forbidden"
	case ErrorTypeTooLong:
		return "Too long"
	case ErrorTypeTooMany:
		return "Too many"
	case ErrorTypeInternal:
		return "Internal error"
	case ErrorTypeTypeInvalid:
		return "Invalid value"
	default:
		panic(fmt.Sprintf("unrecognized validation error: %q", string(t)))
	}
}

func TypeInvalid(field *Path, value interface{}, detail string) *Error {
	return &Error{ErrorTypeTypeInvalid, field.String(), value, detail}
}

func NotFound(field *Path, value interface{}) *Error {
	return &Error{ErrorTypeNotFound, field.String(), value, ""}
}

func Required(field *Path, detail string) *Error {
	return &Error{ErrorTypeRequired, field.String(), "", detail}
}

func Duplicate(field *Path, value interface{}) *Error {
	return &Error{ErrorTypeDuplicate, field.String(), value, ""}
}

func Invalid(field *Path, value interface{}, detail string) *Error {
	return &Error{ErrorTypeInvalid, field.String(), value, detail}
}

func NotSupported(field *Path, value interface{}, validValues []string) *Error {
	detail := ""
	if len(validValues) > 0 {
		quotedValues := make([]string, len(validValues))
		for i, v := range validValues {
			quotedValues[i] = strconv.Quote(v)
		}
		detail = "supported values: " + strings.Join(quotedValues, ", ")
	}
	return &Error{ErrorTypeNotSupported, field.String(), value, detail}
}

func Forbidden(field *Path, detail string) *Error {
	return &Error{ErrorTypeForbidden, field.String(), "", detail}
}

func TooLong(field *Path, value interface{}, maxLength int) *Error {
	return &Error{ErrorTypeTooLong, field.String(), value, fmt.Sprintf("must have at most %d bytes", maxLength)}
}

func TooLongMaxLength(field *Path, value interface{}, maxLength int) *Error {
	var msg string
	if maxLength >= 0 {
		msg = fmt.Sprintf("may not be longer than %d", maxLength)
	} else {
		msg = "value is too long"
	}
	return &Error{ErrorTypeTooLong, field.String(), value, msg}
}

func TooMany(field *Path, actualQuantity, maxQuantity int) *Error {
	var msg string

	if maxQuantity >= 0 {
		msg = fmt.Sprintf("must have at most %d items", maxQuantity)
	} else {
		msg = "has too many items"
	}

	var actual interface{}
	if actualQuantity >= 0 {
		actual = actualQuantity
	} else {
		actual = omitValue
	}

	return &Error{ErrorTypeTooMany, field.String(), actual, msg}
}

func InternalError(field *Path, err error) *Error {
	return &Error{ErrorTypeInternal, field.String(), nil, err.Error()}
}

type ErrorList []*Error

func NewErrorTypeMatcher(t ErrorType) utilerrors.Matcher {
	return func(err error) bool {
		if e, ok := err.(*Error); ok {
			return e.Type == t
		}
		return false
	}
}

func (list ErrorList) ToAggregate() utilerrors.Aggregate {
	if len(list) == 0 {
		return nil
	}
	errs := make([]error, 0, len(list))
	errorMsgs := sets.NewString()
	for _, err := range list {
		msg := fmt.Sprintf("%v", err)
		if errorMsgs.Has(msg) {
			continue
		}
		errorMsgs.Insert(msg)
		errs = append(errs, err)
	}
	return utilerrors.NewAggregate(errs)
}

func fromAggregate(agg utilerrors.Aggregate) ErrorList {
	errs := agg.Errors()
	list := make(ErrorList, len(errs))
	for i := range errs {
		list[i] = errs[i].(*Error)
	}
	return list
}

func (list ErrorList) Filter(fns ...utilerrors.Matcher) ErrorList {
	err := utilerrors.FilterOut(list.ToAggregate(), fns...)
	if err == nil {
		return nil
	}
	return fromAggregate(err.(utilerrors.Aggregate))
}
