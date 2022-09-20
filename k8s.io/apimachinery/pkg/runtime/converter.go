package runtime

import (
	encodingjson "encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/structured-merge-diff/v4/value"

	"k8s.io/apimachinery/pkg/conversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

type UnstructuredConverter interface {
	ToUnstructured(obj interface{}) (map[string]interface{}, error)
	FromUnstructured(u map[string]interface{}, obj interface{}) error
}

type structField struct {
	structType reflect.Type
	field      int
}

type fieldInfo struct {
	name      string
	nameValue reflect.Value
	omitempty bool
}

type fieldsCacheMap map[structField]*fieldInfo

type fieldsCache struct {
	sync.Mutex
	value atomic.Value
}

func newFieldsCache() *fieldsCache {
	cache := &fieldsCache{}
	cache.value.Store(make(fieldsCacheMap))
	return cache
}

var (
	mapStringInterfaceType = reflect.TypeOf(map[string]interface{}{})
	stringType             = reflect.TypeOf(string(""))
	fieldCache             = newFieldsCache()

	DefaultUnstructuredConverter = &unstructuredConverter{
		mismatchDetection: parseBool(os.Getenv("KUBE_PATCH_CONVERSION_DETECTOR")),
		comparison: conversion.EqualitiesOrDie(
			func(a, b time.Time) bool {
				return a.UTC() == b.UTC()
			},
		),
	}
)

func parseBool(key string) bool {
	if len(key) == 0 {
		return false
	}
	value, err := strconv.ParseBool(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't parse '%s' as bool for unstructured mismatch detection", key))
	}
	return value
}

type unstructuredConverter struct {
	mismatchDetection bool
	comparison        conversion.Equalities
}

func NewTestUnstructuredConverter(comparison conversion.Equalities) UnstructuredConverter {
	return NewTestUnstructuredConverterWithValidation(comparison)
}

func NewTestUnstructuredConverterWithValidation(comparison conversion.Equalities) *unstructuredConverter {
	return &unstructuredConverter{
		mismatchDetection: true,
		comparison:        comparison,
	}
}

type fromUnstructuredContext struct {
	isInlined           bool
	matchedKeys         []map[string]struct{}
	parentPath          []string
	returnUnknownFields bool
	unknownFieldErrors  []error
}

func (c *fromUnstructuredContext) pushMatchedKeyTracker() {
	if !c.returnUnknownFields {
		return
	}

	c.matchedKeys = append(c.matchedKeys, nil)
}

func (c *fromUnstructuredContext) recordMatchedKey(key string) {
	if !c.returnUnknownFields {
		return
	}

	last := len(c.matchedKeys) - 1
	if c.matchedKeys[last] == nil {
		c.matchedKeys[last] = map[string]struct{}{}
	}
	c.matchedKeys[last][key] = struct{}{}
}

func (c *fromUnstructuredContext) popAndVerifyMatchedKeys(mapValue reflect.Value) {
	if !c.returnUnknownFields {
		return
	}

	last := len(c.matchedKeys) - 1
	curMatchedKeys := c.matchedKeys[last]
	// 最后一个元素取出来
	c.matchedKeys[last] = nil
	c.matchedKeys = c.matchedKeys[:last]
	for _, key := range mapValue.MapKeys() {
		if _, ok := curMatchedKeys[key.String()]; !ok {
			c.recordUnknownField(key.String())
		}
	}
}

func (c *fromUnstructuredContext) recordUnknownField(field string) {
	if !c.returnUnknownFields {
		return
	}

	pathLen := len(c.parentPath)
	c.pushKey(field)
	errPath := strings.Join(c.parentPath, "")
	c.parentPath = c.parentPath[:pathLen]
	c.unknownFieldErrors = append(c.unknownFieldErrors, fmt.Errorf(`unknown field "%s"`, errPath))
}

func (c *fromUnstructuredContext) pushIndex(index int) {
	if !c.returnUnknownFields {
		return
	}

	c.parentPath = append(c.parentPath, "[", strconv.Itoa(index), "]")
}

func (c *fromUnstructuredContext) pushKey(key string) {
	if !c.returnUnknownFields {
		return
	}

	if len(c.parentPath) > 0 {
		c.parentPath = append(c.parentPath, ".")
	}
	c.parentPath = append(c.parentPath, key)

}

func (c *unstructuredConverter) FromUnstructuredWithValidation(u map[string]interface{}, obj interface{}, returnUnknownFields bool) error {
	t := reflect.TypeOf(obj)
	value := reflect.ValueOf(obj)
	// object 必须是一个指针
	if t.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("FromUnstructured requires a non-nil pointer to an object, got %v", t)
	}

	fromUnstructuredContext := &fromUnstructuredContext{
		returnUnknownFields: returnUnknownFields,
	}
	err := fromUnstructured(reflect.ValueOf(u), value.Elem(), fromUnstructuredContext)
	if c.mismatchDetection {
		// 返回t的元素类型，t必须是数组、切片、chan、map、pointer
		newObj := reflect.New(t.Elem())
		// 将u这个map转换到临时的obj
		newErr := fromUnstructuredViaJSON(u, newObj)
		if (err != nil) != (newErr != nil) {
			klog.Fatalf("FromUnstructured unexpected error for %v: error: %v", u, err)
		}
		if err == nil && !c.comparison.DeepEqual(obj, newObj) {
			klog.Fatalf("FromUnstructured mismatch\nobj1: %#v\nobj2: %#v", obj, newObj)
		}
	}
	if err != nil {
		return err
	}
	if returnUnknownFields && len(fromUnstructuredContext.unknownFieldErrors) > 0 {
		sort.Slice(fromUnstructuredContext.unknownFieldErrors, func(i, j int) bool {
			return fromUnstructuredContext.unknownFieldErrors[i].Error() <
				fromUnstructuredContext.unknownFieldErrors[j].Error()
		})
		return NewStrictDecodingError(fromUnstructuredContext.unknownFieldErrors)
	}
	return nil
}

func (c *unstructuredConverter) FromUnstructured(u map[string]interface{}, obj interface{}) error {
	return c.FromUnstructuredWithValidation(u, obj, false)
}

func fromUnstructuredViaJSON(u map[string]interface{}, obj interface{}) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

func fromUnstructured(sv, dv reflect.Value, ctx *fromUnstructuredContext) error {
	sv = unwrapInterface(sv)
	if !sv.IsValid() {
		dv.Set(reflect.Zero(dv.Type()))
		return nil
	}
	st, dt := sv.Type(), dv.Type()

	switch dt.Kind() {
	case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Struct, reflect.Interface:
	default:
		if st.AssignableTo(dt) {
			dv.Set(sv)
			return nil
		}

		if st.ConvertibleTo(dt) {
			switch st.Kind() {
			case reflect.String:
				switch dt.Kind() {
				case reflect.String:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Bool:
				switch dt.Kind() {
				case reflect.Bool:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				switch dt.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					dv.Set(sv.Convert(dt))
					return nil
				case reflect.Float32, reflect.Float64:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Float32, reflect.Float64:
				switch dt.Kind() {
				case reflect.Float32, reflect.Float64:
					dv.Set(sv.Convert(dt))
					return nil
				}
				if sv.Float() == math.Trunc(sv.Float()) {
					dv.Set(sv.Convert(dt))
					return nil
				}
			}
			return fmt.Errorf("cannot convert %s to %s", st.String(), dt.String())
		}
	}

	// 检查object是否有自定义的json解析器
	entry := value.TypeReflectEntryOf(dv.Type())
	if entry.CanConvertFromUnstructured() {
		return entry.FromUnstructured(sv, dv)
	}

	switch dt.Kind() {
	case reflect.Map:
		return mapFromUnstructured(sv, dv, ctx)
	case reflect.Slice:
		return sliceFromUnstructured(sv, dv, ctx)
	case reflect.Ptr:
		return pointerFromUnstructured(sv, dv, ctx)
	case reflect.Struct:
		return structFromUnstructured(sv, dv, ctx)
	case reflect.Interface:
		return interfaceFromUnstructured(sv, dv)
	default:
		return fmt.Errorf("unrecognized type: %v", dt.Kind())
	}
}

func fieldInfoFromField(structType reflect.Type, field int) *fieldInfo {
	fieldCacheMap := fieldCache.value.Load().(fieldsCacheMap)
	if info, ok := fieldCacheMap[structField{structType, field}]; ok {
		return info
	}

	info := &fieldInfo{}
	typeField := structType.Field(field)
	// 获取这个自动的json标注
	jsonTag := typeField.Tag.Get("json")
	if len(jsonTag) == 0 {
		if typeField.Name == "" {
			info.name = typeField.Name
		} else {
			info.name = strings.ToLower(typeField.Name[:1]) + typeField.Name[1:]
		}
	} else {
		items := strings.Split(jsonTag, ",")
		info.name = items[0]
		for i := range items {
			if items[i] == "omitempty" {
				info.omitempty = true
				break
			}
		}
	}
	info.nameValue = reflect.ValueOf(info.name)

	fieldCache.Lock()
	defer fieldCache.Unlock()
	fieldCacheMap = fieldCache.value.Load().(fieldsCacheMap)
	newFieldCacheMap := make(fieldsCacheMap)
	for k, v := range fieldCacheMap {
		newFieldCacheMap[k] = v
	}
	newFieldCacheMap[structField{structType, field}] = info
	fieldCache.value.Store(newFieldCacheMap)
	return info
}

func unwrapInterface(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func mapFromUnstructured(sv, dv reflect.Value, ctx *fromUnstructuredContext) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() != reflect.Map {
		return fmt.Errorf("cannot restore map from %v", st.Kind())
	}

	if !st.Key().AssignableTo(dt.Key()) && !st.Key().ConvertibleTo(dt.Key()) {
		return fmt.Errorf("cannot copy map with non-assignable keys: %v %v", st.Key(), dt.Key())
	}

	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.MakeMap(dt))
	for _, key := range sv.MapKeys() {
		value := reflect.New(dt.Elem()).Elem()
		if val := unwrapInterface(sv.MapIndex(key)); val.IsValid() {
			if err := fromUnstructured(val, value, ctx); err != nil {
				return err
			}
		} else {
			value.Set(reflect.Zero(dt.Elem()))
		}
		if st.Key().AssignableTo(dt.Key()) {
			dv.SetMapIndex(key, value)
		} else {
			dv.SetMapIndex(key.Convert(dt.Key()), value)
		}
	}
	return nil
}

func sliceFromUnstructured(sv, dv reflect.Value, ctx *fromUnstructuredContext) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() == reflect.String && dt.Elem().Kind() == reflect.Uint8 {
		if len(sv.Interface().(string)) > 0 {
			marshalled, err := json.Marshal(sv.Interface())
			if err != nil {
				return fmt.Errorf("error encoding %s to json: %v", st, err)
			}
			var data []byte
			err = json.Unmarshal(marshalled, &data)
			if err != nil {
				return fmt.Errorf("error decoding from json: %v", err)
			}
			dv.SetBytes(data)
		} else {
			dv.Set(reflect.Zero(dt))
		}
		return nil
	}
	if st.Kind() != reflect.Slice {
		return fmt.Errorf("cannot restore slice from %v", st.Kind())
	}

	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.MakeSlice(dt, sv.Len(), sv.Cap()))

	pathLen := len(ctx.parentPath)
	defer func() {
		ctx.parentPath = ctx.parentPath[:pathLen]
	}()
	for i := 0; i < sv.Len(); i++ {
		ctx.pushIndex(i)
		if err := fromUnstructured(sv.Index(i), dv.Index(i), ctx); err != nil {
			return err
		}
		ctx.parentPath = ctx.parentPath[:pathLen]
	}
	return nil
}

func pointerFromUnstructured(sv, dv reflect.Value, ctx *fromUnstructuredContext) error {
	st, dt := sv.Type(), dv.Type()

	if st.Kind() == reflect.Ptr && sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.New(dt.Elem()))
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		return fromUnstructured(sv.Elem(), dv.Elem(), ctx)
	default:
		return fromUnstructured(sv, dv.Elem(), ctx)
	}
}

func structFromUnstructured(sv, dv reflect.Value, ctx *fromUnstructuredContext) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() != reflect.Map {
		return fmt.Errorf("cannot restore struct from: %v", st.Kind())
	}

	pathLen := len(ctx.parentPath)
	svInlined := ctx.isInlined
	defer func() {
		ctx.parentPath = ctx.parentPath[:pathLen]
		ctx.isInlined = svInlined
	}()
	if !svInlined {
		ctx.pushMatchedKeyTracker()
	}
	for i := 0; i < dt.NumField(); i++ {
		fieldInfo := fieldInfoFromField(dt, i)
		fv := dv.Field(i)

		if len(fieldInfo.name) == 0 {
			ctx.isInlined = true
			if err := fromUnstructured(sv, fv, ctx); err != nil {
				return err
			}
			ctx.isInlined = svInlined
		} else {
			ctx.recordMatchedKey(fieldInfo.name)
			value := unwrapInterface(sv.MapIndex(fieldInfo.nameValue))
			if value.IsValid() {
				ctx.isInlined = false
				ctx.pushKey(fieldInfo.name)
				if err := fromUnstructured(value, fv, ctx); err != nil {
					return err
				}
				ctx.parentPath = ctx.parentPath[:pathLen]
				ctx.isInlined = svInlined
			} else {
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}
	if !svInlined {
		ctx.popAndVerifyMatchedKeys(sv)
	}
	return nil
}

func interfaceFromUnstructured(sv, dv reflect.Value) error {
	// TODO: Is this conversion safe?
	dv.Set(sv)
	return nil
}

func (c *unstructuredConverter) ToUnstructured(obj interface{}) (map[string]interface{}, error) {
	var u map[string]interface{}
	var err error
	if unstr, ok := obj.(Unstructured); ok {
		u = unstr.UnstructuredContent()
	} else {
		t := reflect.TypeOf(obj)
		value := reflect.ValueOf(obj)
		if t.Kind() != reflect.Ptr || value.IsNil() {
			return nil, fmt.Errorf("ToUnstructured requires a non-nil pointer to an object, got %v", t)
		}
		u = map[string]interface{}{}
		err = toUnstructured(value.Elem(), reflect.ValueOf(&u).Elem())
	}
	if c.mismatchDetection {
		newUnstr := map[string]interface{}{}
		newErr := toUnstructuredViaJSON(obj, &newUnstr)
		if (err != nil) != (newErr != nil) {
			klog.Fatalf("ToUnstructured unexpected error for %v: error: %v; newErr: %v", obj, err, newErr)
		}
		if err == nil && !c.comparison.DeepEqual(u, newUnstr) {
			klog.Fatalf("ToUnstructured mismatch\nobj1: %#v\nobj2: %#v", u, newUnstr)
		}
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func DeepCopyJSON(x map[string]interface{}) map[string]interface{} {
	return DeepCopyJSONValue(x).(map[string]interface{})
}

func DeepCopyJSONValue(x interface{}) interface{} {
	switch x := x.(type) {
	case map[string]interface{}:
		if x == nil {
			// Typed nil - an interface{} that contains a type map[string]interface{} with a value of nil
			return x
		}
		clone := make(map[string]interface{}, len(x))
		for k, v := range x {
			clone[k] = DeepCopyJSONValue(v)
		}
		return clone
	case []interface{}:
		if x == nil {
			// Typed nil - an interface{} that contains a type []interface{} with a value of nil
			return x
		}
		clone := make([]interface{}, len(x))
		for i, v := range x {
			clone[i] = DeepCopyJSONValue(v)
		}
		return clone
	case string, int64, bool, float64, nil, encodingjson.Number:
		return x
	default:
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}

func toUnstructuredViaJSON(obj interface{}, u *map[string]interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, u)
}

func toUnstructured(sv, dv reflect.Value) error {
	// Check if the object has a custom string converter.
	entry := value.TypeReflectEntryOf(sv.Type())
	if entry.CanConvertToUnstructured() {
		v, err := entry.ToUnstructured(sv)
		if err != nil {
			return err
		}
		if v != nil {
			dv.Set(reflect.ValueOf(v))
		}
		return nil
	}
	st := sv.Type()
	switch st.Kind() {
	case reflect.String:
		dv.Set(reflect.ValueOf(sv.String()))
		return nil
	case reflect.Bool:
		dv.Set(reflect.ValueOf(sv.Bool()))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dv.Set(reflect.ValueOf(sv.Int()))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uVal := sv.Uint()
		if uVal > math.MaxInt64 {
			return fmt.Errorf("unsigned value %d does not fit into int64 (overflow)", uVal)
		}
		dv.Set(reflect.ValueOf(int64(uVal)))
		return nil
	case reflect.Float32, reflect.Float64:
		dv.Set(reflect.ValueOf(sv.Float()))
		return nil
	case reflect.Map:
		return mapToUnstructured(sv, dv)
	case reflect.Slice:
		return sliceToUnstructured(sv, dv)
	case reflect.Ptr:
		return pointerToUnstructured(sv, dv)
	case reflect.Struct:
		return structToUnstructured(sv, dv)
	case reflect.Interface:
		return interfaceToUnstructured(sv, dv)
	default:
		return fmt.Errorf("unrecognized type: %v", st.Kind())
	}
}

func mapToUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	if dt.Kind() == reflect.Interface && dv.NumMethod() == 0 {
		if st.Key().Kind() == reflect.String {
			dv.Set(reflect.MakeMap(mapStringInterfaceType))
			dv = dv.Elem()
			dt = dv.Type()
		}
	}
	if dt.Kind() != reflect.Map {
		return fmt.Errorf("cannot convert map to: %v", dt.Kind())
	}

	if !st.Key().AssignableTo(dt.Key()) && !st.Key().ConvertibleTo(dt.Key()) {
		return fmt.Errorf("cannot copy map with non-assignable keys: %v %v", st.Key(), dt.Key())
	}

	for _, key := range sv.MapKeys() {
		value := reflect.New(dt.Elem()).Elem()
		if err := toUnstructured(sv.MapIndex(key), value); err != nil {
			return err
		}
		if st.Key().AssignableTo(dt.Key()) {
			dv.SetMapIndex(key, value)
		} else {
			dv.SetMapIndex(key.Convert(dt.Key()), value)
		}
	}
	return nil
}

func sliceToUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	if st.Elem().Kind() == reflect.Uint8 {
		dv.Set(reflect.New(stringType))
		data, err := json.Marshal(sv.Bytes())
		if err != nil {
			return err
		}
		var result string
		if err = json.Unmarshal(data, &result); err != nil {
			return err
		}
		dv.Set(reflect.ValueOf(result))
		return nil
	}
	if dt.Kind() == reflect.Interface && dv.NumMethod() == 0 {
		dv.Set(reflect.MakeSlice(reflect.SliceOf(dt), sv.Len(), sv.Cap()))
		dv = dv.Elem()
		dt = dv.Type()
	}
	if dt.Kind() != reflect.Slice {
		return fmt.Errorf("cannot convert slice to: %v", dt.Kind())
	}
	for i := 0; i < sv.Len(); i++ {
		if err := toUnstructured(sv.Index(i), dv.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func pointerToUnstructured(sv, dv reflect.Value) error {
	if sv.IsNil() {
		// We're done - we don't need to store anything.
		return nil
	}
	return toUnstructured(sv.Elem(), dv)
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Map, reflect.Slice:
		// TODO: It seems that 0-len maps are ignored in it.
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func structToUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if dt.Kind() == reflect.Interface && dv.NumMethod() == 0 {
		dv.Set(reflect.MakeMapWithSize(mapStringInterfaceType, st.NumField()))
		dv = dv.Elem()
		dt = dv.Type()
	}
	if dt.Kind() != reflect.Map {
		return fmt.Errorf("cannot convert struct to: %v", dt.Kind())
	}
	realMap := dv.Interface().(map[string]interface{})

	for i := 0; i < st.NumField(); i++ {
		fieldInfo := fieldInfoFromField(st, i)
		fv := sv.Field(i)

		if fieldInfo.name == "-" {
			// This field should be skipped.
			continue
		}
		if fieldInfo.omitempty && isZero(fv) {
			// omitempty fields should be ignored.
			continue
		}
		if len(fieldInfo.name) == 0 {
			// This field is inlined.
			if err := toUnstructured(fv, dv); err != nil {
				return err
			}
			continue
		}
		switch fv.Type().Kind() {
		case reflect.String:
			realMap[fieldInfo.name] = fv.String()
		case reflect.Bool:
			realMap[fieldInfo.name] = fv.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			realMap[fieldInfo.name] = fv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			realMap[fieldInfo.name] = fv.Uint()
		case reflect.Float32, reflect.Float64:
			realMap[fieldInfo.name] = fv.Float()
		default:
			subv := reflect.New(dt.Elem()).Elem()
			if err := toUnstructured(fv, subv); err != nil {
				return err
			}
			dv.SetMapIndex(fieldInfo.nameValue, subv)
		}
	}
	return nil
}

func interfaceToUnstructured(sv, dv reflect.Value) error {
	if !sv.IsValid() || sv.IsNil() {
		dv.Set(reflect.Zero(dv.Type()))
		return nil
	}
	return toUnstructured(sv.Elem(), dv)
}
