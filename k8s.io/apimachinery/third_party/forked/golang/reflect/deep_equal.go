package reflect

import (
	"fmt"
	"reflect"
	"strings"
)

type Equalities map[reflect.Type]reflect.Value

func EqualitiesOrDie(funcs ...interface{}) Equalities {
	e := Equalities{}
	if err := e.AddFuncs(funcs...); err != nil {
		panic(err)
	}
	return e
}

func (e Equalities) AddFuncs(funcs ...interface{}) error {
	for _, f := range funcs {
		if err := e.AddFunc(f); err != nil {
			return err
		}
	}
	return nil
}

func (e Equalities) AddFunc(eqFunc interface{}) error {
	fv := reflect.ValueOf(eqFunc)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 2 {
		return fmt.Errorf("expected two 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0) != ft.In(1) {
		return fmt.Errorf("expected arg 1 and 2 to have same type, but got %v", ft)
	}
	var forReturnType bool
	boolType := reflect.TypeOf(forReturnType)
	if ft.Out(0) != boolType {
		return fmt.Errorf("expected bool return, got: %v", ft)
	}
	e[ft.In(0)] = fv
	return nil
}

type visit struct {
	a1  uintptr
	a2  uintptr
	typ reflect.Type
}

type unexportedTypePanic []reflect.Type

func (u unexportedTypePanic) Error() string { return u.String() }
func (u unexportedTypePanic) String() string {
	strs := make([]string, len(u))
	for i, t := range u {
		strs[i] = fmt.Sprintf("%v", t)
	}
	return "an unexported field was encountered, nested like this: " + strings.Join(strs, " -> ")
}

func makeUsefulPanic(v reflect.Value) {
	if x := recover(); x != nil {
		if u, ok := x.(unexportedTypePanic); ok {
			u = append(unexportedTypePanic{v.Type()}, u...)
			x = u
		}
		panic(x)
	}
}

func (e Equalities) deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	defer makeUsefulPanic(v1)

	// 如果值为零值，就表示无效
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	// 说明类型不同的了零值是相等的
	if v1.Type() != v2.Type() {
		return false
	}
	// 找到了直接作为函数调用
	if fv, ok := e[v1.Type()]; ok {
		return fv.Call([]reflect.Value{v1, v2})[0].Bool()
	}

	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := v1.UnsafeAddr()
		addr2 := v2.UnsafeAddr()
		if addr1 > addr2 {
			addr1, addr2 = addr2, addr1
		}

		if addr1 == addr2 {
			return true
		}

		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return false
		}
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return e.deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		return e.deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !e.deepValueEqual(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return false
		}
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			if !e.deepValueEqual(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	default:
		if !v1.CanInterface() || !v2.CanInterface() {
			panic(unexportedTypePanic{})
		}
		return v1.Interface() == v2.Interface()
	}
}

func (e Equalities) DeepEqual(a1, a2 interface{}) bool {
	if a1 == nil || a2 == nil {
		return a1 == a2
	}
	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)
	if v1.Type() != v2.Type() {
		return false
	}
	return e.deepValueEqual(v1, v2, make(map[visit]bool), 0)
}

func (e Equalities) deepValueDerive(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	defer makeUsefulPanic(v1)

	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}
	if fv, ok := e[v1.Type()]; ok {
		return fv.Call([]reflect.Value{v1, v2})[0].Bool()
	}

	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := v1.UnsafeAddr()
		addr2 := v2.UnsafeAddr()
		if addr1 > addr2 {
			// Canonicalize order to reduce number of entries in visited.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are identical ...
		if addr1 == addr2 {
			return true
		}

		// ... or already seen
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		// We don't need to check length here because length is part of
		// an array's type, which has already been filtered for.
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueDerive(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() > v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !e.deepValueDerive(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.String:
		if v1.Len() == 0 {
			return true
		}
		if v1.Len() > v2.Len() {
			return false
		}
		return v1.String() == v2.String()
	case reflect.Interface:
		if v1.IsNil() {
			return true
		}
		return e.deepValueDerive(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		if v1.IsNil() {
			return true
		}
		return e.deepValueDerive(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !e.deepValueDerive(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v1.IsNil() || v1.Len() == 0 {
			return true
		}
		if v1.Len() > v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			if !e.deepValueDerive(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	default:
		// Normal equality suffices
		if !v1.CanInterface() || !v2.CanInterface() {
			panic(unexportedTypePanic{})
		}
		return v1.Interface() == v2.Interface()
	}
}

func (e Equalities) DeepDerivative(a1, a2 interface{}) bool {
	if a1 == nil {
		return true
	}
	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)
	if v1.Type() != v2.Type() {
		return false
	}
	return e.deepValueDerive(v1, v2, make(map[visit]bool), 0)
}
