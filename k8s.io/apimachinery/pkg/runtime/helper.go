package runtime

import (
	"fmt"
	"io"
	"reflect"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
)

type unsafeObjectConvertor struct {
	*Scheme
}

var _ ObjectConvertor = unsafeObjectConvertor{}

func (c unsafeObjectConvertor) ConvertToVersion(in Object, outVersion GroupVersioner) (Object, error) {
	return c.Scheme.UnsafeConvertToVersion(in, outVersion)
}

func UnsafeObjectConvertor(scheme *Scheme) ObjectConvertor {
	return unsafeObjectConvertor{scheme}
}

func SetField(src interface{}, v reflect.Value, fieldName string) error {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("couldn't find %v field in %T", fieldName, v.Interface())
	}
	srcValue := reflect.ValueOf(src)
	if srcValue.Type().AssignableTo(field.Type()) {
		field.Set(srcValue)
		return nil
	}
	if srcValue.Type().ConvertibleTo(field.Type()) {
		field.Set(srcValue.Convert(field.Type()))
		return nil
	}
	return fmt.Errorf("couldn't assign/convert %v to %v", srcValue.Type(), field.Type())
}

func Field(v reflect.Value, fieldName string, dest interface{}) error {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("couldn't find %v field in %T", fieldName, v.Interface())
	}
	destValue, err := conversion.EnforcePtr(dest)
	if err != nil {
		return err
	}
	if field.Type().AssignableTo(destValue.Type()) {
		destValue.Set(field)
		return nil
	}
	if field.Type().ConvertibleTo(destValue.Type()) {
		destValue.Set(field.Convert(destValue.Type()))
		return nil
	}
	return fmt.Errorf("couldn't assign/convert %v to %v", field.Type(), destValue.Type())
}

func FieldPtr(v reflect.Value, fieldName string, dest interface{}) error {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("couldn't find %v field in %T", fieldName, v.Interface())
	}
	v, err := conversion.EnforcePtr(dest)
	if err != nil {
		return err
	}
	field = field.Addr()
	if field.Type().AssignableTo(v.Type()) {
		v.Set(field)
		return nil
	}
	if field.Type().ConvertibleTo(v.Type()) {
		v.Set(field.Convert(v.Type()))
		return nil
	}
	return fmt.Errorf("couldn't assign/convert %v to %v", field.Type(), v.Type())
}

func EncodeList(e Encoder, objects []Object) error {
	var errs []error
	for i := range objects {
		data, err := Encode(e, objects[i])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// TODO: Set ContentEncoding and ContentType.
		objects[i] = &Unknown{Raw: data}
	}
	return errors.NewAggregate(errs)
}

func decodeListItem(obj *Unknown, decoders []Decoder) (Object, error) {
	for _, decoder := range decoders {
		// TODO: Decode based on ContentType.
		obj, err := Decode(decoder, obj.Raw)
		if err != nil {
			if IsNotRegisteredError(err) {
				continue
			}
			return nil, err
		}
		return obj, nil
	}
	// could not decode, so leave the object as Unknown, but give the decoders the
	// chance to set Unknown.TypeMeta if it is available.
	for _, decoder := range decoders {
		if err := DecodeInto(decoder, obj.Raw, obj); err == nil {
			return obj, nil
		}
	}
	return obj, nil
}

func DecodeList(objects []Object, decoders ...Decoder) []error {
	errs := []error(nil)
	for i, obj := range objects {
		switch t := obj.(type) {
		case *Unknown:
			decoded, err := decodeListItem(t, decoders)
			if err != nil {
				errs = append(errs, err)
				break
			}
			objects[i] = decoded
		}
	}
	return errs
}

type MultiObjectTyper []ObjectTyper

var _ ObjectTyper = MultiObjectTyper{}

func (m MultiObjectTyper) ObjectKinds(obj Object) (gvks []schema.GroupVersionKind, unversionedType bool, err error) {
	for _, t := range m {
		gvks, unversionedType, err = t.ObjectKinds(obj)
		if err == nil {
			return
		}
	}
	return
}

func (m MultiObjectTyper) Recognizes(gvk schema.GroupVersionKind) bool {
	for _, t := range m {
		if t.Recognizes(gvk) {
			return true
		}
	}
	return false
}

func SetZeroValue(objPtr Object) error {
	v, err := conversion.EnforcePtr(objPtr)
	if err != nil {
		return err
	}
	v.Set(reflect.Zero(v.Type()))
	return nil
}

var DefaultFramer = defaultFramer{}

type defaultFramer struct{}

func (defaultFramer) NewFrameReader(r io.ReadCloser) io.ReadCloser { return r }
func (defaultFramer) NewFrameWriter(w io.Writer) io.Writer         { return w }

type WithVersionEncoder struct {
	Version GroupVersioner
	Encoder
	ObjectTyper
}

func (e WithVersionEncoder) Encode(obj Object, stream io.Writer) error {
	gvks, _, err := e.ObjectTyper.ObjectKinds(obj)
	if err != nil {
		if IsNotRegisteredError(err) {
			return e.Encoder.Encode(obj, stream)
		}
		return err
	}
	kind := obj.GetObjectKind()
	oldGVK := kind.GroupVersionKind()
	gvk := gvks[0]
	if e.Version != nil {
		preferredGVK, ok := e.Version.KindForGroupVersionKinds(gvks)
		if ok {
			gvk = preferredGVK
		}
	}
	kind.SetGroupVersionKind(gvk)
	err = e.Encoder.Encode(obj, stream)
	kind.SetGroupVersionKind(oldGVK)
	return err
}

type WithoutVersionDecoder struct {
	Decoder
}

// Decode does not do conversion. It removes the gvk during deserialization.
func (d WithoutVersionDecoder) Decode(data []byte, defaults *schema.GroupVersionKind, into Object) (Object, *schema.GroupVersionKind, error) {
	obj, gvk, err := d.Decoder.Decode(data, defaults, into)
	if obj != nil {
		kind := obj.GetObjectKind()
		// clearing the gvk is just a convention of a codec
		kind.SetGroupVersionKind(schema.GroupVersionKind{})
	}
	return obj, gvk, err
}
