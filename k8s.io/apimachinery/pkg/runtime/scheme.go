package runtime

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/naming"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Scheme struct {
	gvkToType                 map[schema.GroupVersionKind]reflect.Type
	typeToGVK                 map[reflect.Type][]schema.GroupVersionKind
	unversionedTypes          map[reflect.Type]schema.GroupVersionKind
	unversionedKinds          map[string]reflect.Type
	fieldLabelConversionFuncs map[schema.GroupVersionKind]FieldLabelConversionFunc
	defaulterFuncs            map[reflect.Type]func(interface{})
	converter                 *conversion.Converter
	versionPriority           map[string][]string
	observedVersions          []schema.GroupVersion
	schemeName                string
}

type FieldLabelConversionFunc func(label, value string) (internalLabel, internalValue string, err error)

// p93
func NewScheme() *Scheme {
	s := &Scheme{
		gvkToType:                 map[schema.GroupVersionKind]reflect.Type{},
		typeToGVK:                 map[reflect.Type][]schema.GroupVersionKind{},
		unversionedTypes:          map[reflect.Type]schema.GroupVersionKind{},
		unversionedKinds:          map[string]reflect.Type{},
		fieldLabelConversionFuncs: map[schema.GroupVersionKind]FieldLabelConversionFunc{},
		defaulterFuncs:            map[reflect.Type]func(interface{}){},
		versionPriority:           map[string][]string{},
		schemeName:                naming.GetNameFromCallsite(internalPackages...),
	}
	s.converter = conversion.NewConverter(nil)
	utilruntime.Must(RegisterEmbeddedConversions(s))
	utilruntime.Must(RegisterStringConversions(s))
	return s
}

func (s *Scheme) Converter() *conversion.Converter {
	return s.converter
}

func (s *Scheme) AddUnversionedTypes(version schema.GroupVersion, types ...Object) {
	s.addObservedVersion(version)
	s.AddKnownTypes(version, types...)
	for _, obj := range types {
		t := reflect.TypeOf(obj).Elem()
		gvk := version.WithKind(t.Name())
		s.unversionedTypes[t] = gvk
		if old, ok := s.unversionedKinds[gvk.Kind]; ok && t != old {
			panic(fmt.Sprintf("%v.%v has already been registered as unversioned kind %q - kind name must be unique in scheme %q", old.PkgPath(), old.Name(), gvk, s.schemeName))
		}
		s.unversionedKinds[gvk.Kind] = t
	}
}

func (s *Scheme) AddKnownTypes(gv schema.GroupVersion, types ...Object) {
	s.addObservedVersion(gv)
	for _, obj := range types {
		t := reflect.TypeOf(obj)
		// 必须是指针
		if t.Kind() != reflect.Ptr {
			panic("All types must be pointers to structs.")
		}
		// 拿到元素
		t = t.Elem()
		s.AddKnownTypeWithName(gv.WithKind(t.Name()), obj)
	}
}

func (s *Scheme) AddKnownTypeWithName(gvk schema.GroupVersionKind, obj Object) {
	s.addObservedVersion(gvk.GroupVersion())
	t := reflect.TypeOf(obj)
	if len(gvk.Version) == 0 {
		panic(fmt.Sprintf("version is required on all types: %s %v", gvk, t))
	}
	if t.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs.")
	}
	t = t.Elem()
	// 必须是结构体
	if t.Kind() != reflect.Struct {
		panic("All types must be pointers to structs.")
	}

	if oldT, found := s.gvkToType[gvk]; found && oldT != t {
		panic(fmt.Sprintf("Double registration of different types for %v: old=%v.%v, new=%v.%v in scheme %q", gvk, oldT.PkgPath(), oldT.Name(), t.PkgPath(), t.Name(), s.schemeName))
	}
	// 一个确定gvk只能转固定的内部类型
	s.gvkToType[gvk] = t

	for _, existingGvk := range s.typeToGVK[t] {
		if existingGvk == gvk {
			return
		}
	}
	// 一个类型是内部类型，是可以转不同的外部类型的，所以是数组
	s.typeToGVK[t] = append(s.typeToGVK[t], gvk)

	// 只能有一个输入参数没有输出参数，且输入参数必须是与obj相同的
	if m := reflect.ValueOf(obj).MethodByName("DeepCopyInto"); m.IsValid() && m.Type().NumIn() == 1 && m.Type().NumOut() == 0 && m.Type().In(0) == reflect.TypeOf(obj) {
		if err := s.AddGeneratedConversionFunc(obj, obj, func(a, b interface{}, scope conversion.Scope) error {
			// copy a to b
			reflect.ValueOf(a).MethodByName("DeepCopyInto").Call([]reflect.Value{reflect.ValueOf(b)})
			// Object对应的结构体都有TypeMeta字段，GetObjectKind会返回其内嵌的TypeMeta成员
			// 调用其SetGroupVersionKind会设置TypeMeta的APIVersion和Kind字段
			// APVersion为 GROUP/VERSION; Kind为 KIND
			b.(Object).GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
			return nil
		}); err != nil {
			panic(err)
		}
	}
}

func (s *Scheme) KnownTypes(gv schema.GroupVersion) map[string]reflect.Type {
	types := make(map[string]reflect.Type)
	for gvk, t := range s.gvkToType {
		if gv != gvk.GroupVersion() {
			continue
		}

		types[gvk.Kind] = t
	}
	return types
}

func (s *Scheme) VersionsForGroupKind(gk schema.GroupKind) []schema.GroupVersion {
	availableVersions := []schema.GroupVersion{}
	for gvk := range s.gvkToType {
		if gk != gvk.GroupKind() {
			continue
		}

		availableVersions = append(availableVersions, gvk.GroupVersion())
	}

	// order the return for stability
	ret := []schema.GroupVersion{}
	for _, version := range s.PrioritizedVersionsForGroup(gk.Group) {
		for _, availableVersion := range availableVersions {
			if version != availableVersion {
				continue
			}
			ret = append(ret, availableVersion)
		}
	}

	return ret
}

func (s *Scheme) AllKnownTypes() map[schema.GroupVersionKind]reflect.Type {
	return s.gvkToType
}

func (s *Scheme) ObjectKinds(obj Object) ([]schema.GroupVersionKind, bool, error) {
	// Unstructured objects are always considered to have their declared GVK
	if _, ok := obj.(Unstructured); ok {
		// we require that the GVK be populated in order to recognize the object
		gvk := obj.GetObjectKind().GroupVersionKind()
		if len(gvk.Kind) == 0 {
			return nil, false, NewMissingKindErr("unstructured object has no kind")
		}
		if len(gvk.Version) == 0 {
			return nil, false, NewMissingVersionErr("unstructured object has no version")
		}
		return []schema.GroupVersionKind{gvk}, false, nil
	}

	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return nil, false, err
	}
	t := v.Type()

	gvks, ok := s.typeToGVK[t]
	if !ok {
		return nil, false, NewNotRegisteredErrForType(s.schemeName, t)
	}
	_, unversionedType := s.unversionedTypes[t]

	return gvks, unversionedType, nil
}

func (s *Scheme) Recognizes(gvk schema.GroupVersionKind) bool {
	_, exists := s.gvkToType[gvk]
	return exists
}

func (s *Scheme) IsUnversioned(obj Object) (bool, bool) {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return false, false
	}
	t := v.Type()

	if _, ok := s.typeToGVK[t]; !ok {
		return false, false
	}
	_, ok := s.unversionedTypes[t]
	return ok, true
}

func (s *Scheme) New(kind schema.GroupVersionKind) (Object, error) {
	if t, exists := s.gvkToType[kind]; exists {
		return reflect.New(t).Interface().(Object), nil
	}

	if t, exists := s.unversionedKinds[kind.Kind]; exists {
		return reflect.New(t).Interface().(Object), nil
	}
	return nil, NewNotRegisteredErrForKind(s.schemeName, kind)
}

func (s *Scheme) AddIgnoredConversionType(from, to interface{}) error {
	return s.converter.RegisterIgnoredConversion(from, to)
}

// p318
func (s *Scheme) AddConversionFunc(a, b interface{}, fn conversion.ConversionFunc) error {
	return s.converter.RegisterUntypedConversionFunc(a, b, fn)
}

func (s *Scheme) AddGeneratedConversionFunc(a, b interface{}, fn conversion.ConversionFunc) error {
	return s.converter.RegisterGeneratedUntypedConversionFunc(a, b, fn)
}

func (s *Scheme) AddFieldLabelConversionFunc(gvk schema.GroupVersionKind, conversionFunc FieldLabelConversionFunc) error {
	s.fieldLabelConversionFuncs[gvk] = conversionFunc
	return nil
}

func (s *Scheme) AddTypeDefaultingFunc(srcType Object, fn func(interface{})) {
	s.defaulterFuncs[reflect.TypeOf(srcType)] = fn
}

func (s *Scheme) Default(src Object) {
	if fn, ok := s.defaulterFuncs[reflect.TypeOf(src)]; ok {
		fn(src)
	}
}

func (s *Scheme) Convert(in, out interface{}, context interface{}) error {
	unstructuredIn, okIn := in.(Unstructured)
	unstructuredOut, okOut := out.(Unstructured)
	switch {
	case okIn && okOut:
		unstructuredOut.SetUnstructuredContent(unstructuredIn.UnstructuredContent())
		return nil
	case okOut:
		obj, ok := in.(Object)
		if !ok {
			return fmt.Errorf("unable to convert object type %T to Unstructured, must be a runtime.Object", in)
		}
		gvks, unversioned, err := s.ObjectKinds(obj)
		if err != nil {
			return err
		}
		gvk := gvks[0]

		if unversioned || gvk.Version != APIVersionInternal {
			content, err := DefaultUnstructuredConverter.ToUnstructured(in)
			if err != nil {
				return err
			}
			unstructuredOut.SetUnstructuredContent(content)
			unstructuredOut.GetObjectKind().SetGroupVersionKind(gvk)
			return nil
		}

		// attempt to convert the object to an external version first.
		target, ok := context.(GroupVersioner)
		if !ok {
			return fmt.Errorf("unable to convert the internal object type %T to Unstructured without providing a preferred version to convert to", in)
		}
		// Convert is implicitly unsafe, so we don't need to perform a safe conversion
		versioned, err := s.UnsafeConvertToVersion(obj, target)
		if err != nil {
			return err
		}
		content, err := DefaultUnstructuredConverter.ToUnstructured(versioned)
		if err != nil {
			return err
		}
		unstructuredOut.SetUnstructuredContent(content)
		return nil
	case okIn:
		typed, err := s.unstructuredToTyped(unstructuredIn)
		if err != nil {
			return err
		}
		in = typed
	}

	meta := s.generateConvertMeta(in)
	meta.Context = context
	return s.converter.Convert(in, out, meta)
}

func (s *Scheme) ConvertFieldLabel(gvk schema.GroupVersionKind, label, value string) (string, string, error) {
	conversionFunc, ok := s.fieldLabelConversionFuncs[gvk]
	if !ok {
		return DefaultMetaV1FieldSelectorConversion(label, value)
	}
	return conversionFunc(label, value)
}

func (s *Scheme) ConvertToVersion(in Object, target GroupVersioner) (Object, error) {
	return s.convertToVersion(true, in, target)
}

func (s *Scheme) UnsafeConvertToVersion(in Object, target GroupVersioner) (Object, error) {
	return s.convertToVersion(false, in, target)
}

func (s *Scheme) convertToVersion(copy bool, in Object, target GroupVersioner) (Object, error) {
	var t reflect.Type

	if u, ok := in.(Unstructured); ok {
		typed, err := s.unstructuredToTyped(u)
		if err != nil {
			return nil, err
		}

		in = typed
		// unstructuredToTyped returns an Object, which must be a pointer to a struct.
		t = reflect.TypeOf(in).Elem()

	} else {
		// determine the incoming kinds with as few allocations as possible.
		t = reflect.TypeOf(in)
		if t.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("only pointer types may be converted: %v", t)
		}
		t = t.Elem()
		if t.Kind() != reflect.Struct {
			return nil, fmt.Errorf("only pointers to struct types may be converted: %v", t)
		}
	}

	kinds, ok := s.typeToGVK[t]
	if !ok || len(kinds) == 0 {
		return nil, NewNotRegisteredErrForType(s.schemeName, t)
	}

	gvk, ok := target.KindForGroupVersionKinds(kinds)
	if !ok {
		// try to see if this type is listed as unversioned (for legacy support)
		// TODO: when we move to server API versions, we should completely remove the unversioned concept
		if unversionedKind, ok := s.unversionedTypes[t]; ok {
			if gvk, ok := target.KindForGroupVersionKinds([]schema.GroupVersionKind{unversionedKind}); ok {
				return copyAndSetTargetKind(copy, in, gvk)
			}
			return copyAndSetTargetKind(copy, in, unversionedKind)
		}
		return nil, NewNotRegisteredErrForTarget(s.schemeName, t, target)
	}

	// target wants to use the existing type, set kind and return (no conversion necessary)
	for _, kind := range kinds {
		if gvk == kind {
			return copyAndSetTargetKind(copy, in, gvk)
		}
	}

	// type is unversioned, no conversion necessary
	if unversionedKind, ok := s.unversionedTypes[t]; ok {
		if gvk, ok := target.KindForGroupVersionKinds([]schema.GroupVersionKind{unversionedKind}); ok {
			return copyAndSetTargetKind(copy, in, gvk)
		}
		return copyAndSetTargetKind(copy, in, unversionedKind)
	}

	out, err := s.New(gvk)
	if err != nil {
		return nil, err
	}

	if copy {
		in = in.DeepCopyObject()
	}

	meta := s.generateConvertMeta(in)
	meta.Context = target
	if err := s.converter.Convert(in, out, meta); err != nil {
		return nil, err
	}

	setTargetKind(out, gvk)
	return out, nil
}

func (s *Scheme) unstructuredToTyped(in Unstructured) (Object, error) {
	// the type must be something we recognize
	gvks, _, err := s.ObjectKinds(in)
	if err != nil {
		return nil, err
	}
	typed, err := s.New(gvks[0])
	if err != nil {
		return nil, err
	}
	if err := DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), typed); err != nil {
		return nil, fmt.Errorf("unable to convert unstructured object to %v: %v", gvks[0], err)
	}
	return typed, nil
}

func (s *Scheme) generateConvertMeta(in interface{}) *conversion.Meta {
	return s.converter.DefaultMeta(reflect.TypeOf(in))
}

func copyAndSetTargetKind(copy bool, obj Object, kind schema.GroupVersionKind) (Object, error) {
	if copy {
		obj = obj.DeepCopyObject()
	}
	setTargetKind(obj, kind)
	return obj, nil
}

func setTargetKind(obj Object, kind schema.GroupVersionKind) {
	if kind.Version == APIVersionInternal {
		// internal is a special case
		// TODO: look at removing the need to special case this
		obj.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
		return
	}
	obj.GetObjectKind().SetGroupVersionKind(kind)
}

func (s *Scheme) SetVersionPriority(versions ...schema.GroupVersion) error {
	groups := sets.String{}
	order := []string{}
	for _, version := range versions {
		if len(version.Version) == 0 || version.Version == APIVersionInternal {
			return fmt.Errorf("internal versions cannot be prioritized: %v", version)
		}

		groups.Insert(version.Group)
		order = append(order, version.Version)
	}
	if len(groups) != 1 {
		return fmt.Errorf("must register versions for exactly one group: %v", strings.Join(groups.List(), ", "))
	}

	// 期望的是穿甲的version的group都是一样，所以取第一个就行
	s.versionPriority[groups.List()[0]] = order
	return nil
}

func (s *Scheme) PrioritizedVersionsForGroup(group string) []schema.GroupVersion {
	ret := []schema.GroupVersion{}
	for _, version := range s.versionPriority[group] {
		ret = append(ret, schema.GroupVersion{Group: group, Version: version})
	}
	for _, observedVersion := range s.observedVersions {
		if observedVersion.Group != group {
			continue
		}
		found := false
		for _, existing := range ret {
			if existing == observedVersion {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, observedVersion)
		}
	}

	return ret
}

func (s *Scheme) PrioritizedVersionsAllGroups() []schema.GroupVersion {
	ret := []schema.GroupVersion{}
	for group, versions := range s.versionPriority {
		for _, version := range versions {
			ret = append(ret, schema.GroupVersion{Group: group, Version: version})
		}
	}
	for _, observedVersion := range s.observedVersions {
		found := false
		for _, existing := range ret {
			if existing == observedVersion {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, observedVersion)
		}
	}
	return ret
}

func (s *Scheme) PreferredVersionAllGroups() []schema.GroupVersion {
	ret := []schema.GroupVersion{}
	for group, versions := range s.versionPriority {
		for _, version := range versions {
			ret = append(ret, schema.GroupVersion{Group: group, Version: version})
			break
		}
	}
	for _, observedVersion := range s.observedVersions {
		found := false
		for _, existing := range ret {
			if existing.Group == observedVersion.Group {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, observedVersion)
		}
	}

	return ret
}

func (s *Scheme) IsGroupRegistered(group string) bool {
	for _, observedVersion := range s.observedVersions {
		if observedVersion.Group == group {
			return true
		}
	}
	return false
}

func (s *Scheme) IsVersionRegistered(version schema.GroupVersion) bool {
	for _, observedVersion := range s.observedVersions {
		if observedVersion == version {
			return true
		}
	}

	return false
}

// p689
func (s *Scheme) addObservedVersion(version schema.GroupVersion) {
	if len(version.Version) == 0 || version.Version == APIVersionInternal {
		return
	}
	for _, observedVersion := range s.observedVersions {
		if observedVersion == version {
			return
		}
	}
	s.observedVersions = append(s.observedVersions, version)
}

func (s *Scheme) Name() string {
	return s.schemeName
}

var internalPackages = []string{"k8s.io/apimachinery/pkg/runtime/scheme.go"}
