package conversion

import (
	"fmt"
	"reflect"
)

type typePair struct {
	source reflect.Type
	dest   reflect.Type
}

type NameFunc func(t reflect.Type) string

var DefaultNameFunc = func(t reflect.Type) string { return t.Name() }

type ConversionFunc func(a, b interface{}, scope Scope) error

type Converter struct {
	conversionFuncs           ConversionFuncs
	generatedConversionFuncs  ConversionFuncs
	ignoredUntypedConversions map[typePair]struct{}
}

func NewConverter(NameFunc) *Converter {
	c := &Converter{
		conversionFuncs:           NewConversionFuncs(),
		generatedConversionFuncs:  NewConversionFuncs(),
		ignoredUntypedConversions: make(map[typePair]struct{}),
	}
	c.RegisterUntypedConversionFunc(
		(*[]byte)(nil), (*[]byte)(nil),
		func(a, b interface{}, s Scope) error {
			return Convert_Slice_byte_To_Slice_byte(a.(*[]byte), b.(*[]byte), s)
		},
	)
	return c
}

func (c *Converter) WithConversions(fns ConversionFuncs) *Converter {
	copied := *c
	copied.conversionFuncs = c.conversionFuncs.Merge(fns)
	return &copied
}

func (c *Converter) DefaultMeta(t reflect.Type) *Meta {
	return &Meta{}
}

func Convert_Slice_byte_To_Slice_byte(in *[]byte, out *[]byte, s Scope) error {
	if *in == nil {
		*out = nil
		return nil
	}
	*out = make([]byte, len(*in))
	copy(*out, *in)
	return nil
}

type Scope interface {
	Convert(src, dest interface{}) error
	Meta() *Meta
}

func NewConversionFuncs() ConversionFuncs {
	return ConversionFuncs{
		untyped: make(map[typePair]ConversionFunc),
	}
}

type ConversionFuncs struct {
	untyped map[typePair]ConversionFunc
}

func (c ConversionFuncs) AddUntyped(a, b interface{}, fn ConversionFunc) error {
	tA, tB := reflect.TypeOf(a), reflect.TypeOf(b)
	if tA.Kind() != reflect.Ptr {
		return fmt.Errorf("the type %T must be a pointer to register as an untyped conversion", a)
	}
	if tB.Kind() != reflect.Ptr {
		return fmt.Errorf("the type %T must be a pointer to register as an untyped conversion", b)
	}
	// fmt.Printf("xxxxxx [%s] [%s]\n", tA.Elem(), tB.Elem())
	c.untyped[typePair{tA, tB}] = fn
	return nil
}

func (c ConversionFuncs) Merge(other ConversionFuncs) ConversionFuncs {
	merged := NewConversionFuncs()
	for k, v := range c.untyped {
		merged.untyped[k] = v
	}
	for k, v := range other.untyped {
		merged.untyped[k] = v
	}
	return merged
}

type Meta struct {
	Context interface{}
}

type scope struct {
	converter *Converter
	meta      *Meta
}

func (s *scope) Convert(src, dest interface{}) error {
	return s.converter.Convert(src, dest, s.meta)
}

func (s *scope) Meta() *Meta {
	return s.meta
}

func (c *Converter) RegisterUntypedConversionFunc(a, b interface{}, fn ConversionFunc) error {
	return c.conversionFuncs.AddUntyped(a, b, fn)
}

func (c *Converter) RegisterGeneratedUntypedConversionFunc(a, b interface{}, fn ConversionFunc) error {
	return c.generatedConversionFuncs.AddUntyped(a, b, fn)
}

func (c *Converter) RegisterIgnoredConversion(from, to interface{}) error {
	typeFrom := reflect.TypeOf(from)
	typeTo := reflect.TypeOf(to)
	if typeFrom.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'from' param 0, got: %v", typeFrom)
	}
	if typeTo.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'to' param 1, got: %v", typeTo)
	}
	c.ignoredUntypedConversions[typePair{typeFrom, typeTo}] = struct{}{}
	return nil
}

func (c *Converter) Convert(src, dest interface{}, meta *Meta) error {
	pair := typePair{reflect.TypeOf(src), reflect.TypeOf(dest)}
	scope := &scope{
		converter: c,
		meta:      meta,
	}

	// 如果被忽略了，就啥也不做
	if _, ok := c.ignoredUntypedConversions[pair]; ok {
		return nil
	}
	// 先调用转换函数
	if fn, ok := c.conversionFuncs.untyped[pair]; ok {
		return fn(src, dest, scope)
	}
	// 没有就调用生成的转换函数
	if fn, ok := c.generatedConversionFuncs.untyped[pair]; ok {
		return fn(src, dest, scope)
	}
	dv, err := EnforcePtr(dest)
	if err != nil {
		return err
	}
	sv, err := EnforcePtr(src)
	if err != nil {
		return err
	}
	return fmt.Errorf("converting (%s) to (%s): unknown conversion", sv.Type(), dv.Type())
}
