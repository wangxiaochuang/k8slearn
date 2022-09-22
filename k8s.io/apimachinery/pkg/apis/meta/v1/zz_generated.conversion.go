package v1

import (
	"net/url"
	"unsafe"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*CreateOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_CreateOptions(a.(*url.Values), b.(*CreateOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*DeleteOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_DeleteOptions(a.(*url.Values), b.(*DeleteOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*GetOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_GetOptions(a.(*url.Values), b.(*GetOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*ListOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_ListOptions(a.(*url.Values), b.(*ListOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*PatchOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_PatchOptions(a.(*url.Values), b.(*PatchOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*TableOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_TableOptions(a.(*url.Values), b.(*TableOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*UpdateOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_UpdateOptions(a.(*url.Values), b.(*UpdateOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*map[string]string)(nil), (*LabelSelector)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Map_string_To_string_To_v1_LabelSelector(a.(*map[string]string), b.(*LabelSelector), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**bool)(nil), (*bool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_bool_To_bool(a.(**bool), b.(*bool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**float64)(nil), (*float64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_float64_To_float64(a.(**float64), b.(*float64), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**int32)(nil), (*int32)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_int32_To_int32(a.(**int32), b.(*int32), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**int64)(nil), (*int)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_int64_To_int(a.(**int64), b.(*int), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**int64)(nil), (*int64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_int64_To_int64(a.(**int64), b.(*int64), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**intstr.IntOrString)(nil), (*intstr.IntOrString)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_intstr_IntOrString_To_intstr_IntOrString(a.(**intstr.IntOrString), b.(*intstr.IntOrString), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**string)(nil), (*string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_string_To_string(a.(**string), b.(*string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((**Duration)(nil), (*Duration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Pointer_v1_Duration_To_v1_Duration(a.(**Duration), b.(*Duration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (**DeletionPropagation)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_Pointer_v1_DeletionPropagation(a.(*[]string), b.(**DeletionPropagation), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (**Time)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_Pointer_v1_Time(a.(*[]string), b.(**Time), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*[]int32)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_Slice_int32(a.(*[]string), b.(*[]int32), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*IncludeObjectPolicy)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_v1_IncludeObjectPolicy(a.(*[]string), b.(*IncludeObjectPolicy), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*ResourceVersionMatch)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_v1_ResourceVersionMatch(a.(*[]string), b.(*ResourceVersionMatch), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*Time)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_v1_Time(a.(*[]string), b.(*Time), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*bool)(nil), (**bool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_bool_To_Pointer_bool(a.(*bool), b.(**bool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*fields.Selector)(nil), (*string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_fields_Selector_To_string(a.(*fields.Selector), b.(*string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*float64)(nil), (**float64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_float64_To_Pointer_float64(a.(*float64), b.(**float64), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*int32)(nil), (**int32)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_int32_To_Pointer_int32(a.(*int32), b.(**int32), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*int64)(nil), (**int64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_int64_To_Pointer_int64(a.(*int64), b.(**int64), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*int)(nil), (**int64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_int_To_Pointer_int64(a.(*int), b.(**int64), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*intstr.IntOrString)(nil), (**intstr.IntOrString)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_intstr_IntOrString_To_Pointer_intstr_IntOrString(a.(*intstr.IntOrString), b.(**intstr.IntOrString), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*intstr.IntOrString)(nil), (*intstr.IntOrString)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_intstr_IntOrString_To_intstr_IntOrString(a.(*intstr.IntOrString), b.(*intstr.IntOrString), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*labels.Selector)(nil), (*string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_labels_Selector_To_string(a.(*labels.Selector), b.(*string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*resource.Quantity)(nil), (*resource.Quantity)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_resource_Quantity_To_resource_Quantity(a.(*resource.Quantity), b.(*resource.Quantity), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*string)(nil), (**string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_string_To_Pointer_string(a.(*string), b.(**string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*string)(nil), (*fields.Selector)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_string_To_fields_Selector(a.(*string), b.(*fields.Selector), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*string)(nil), (*labels.Selector)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_string_To_labels_Selector(a.(*string), b.(*labels.Selector), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*url.Values)(nil), (*DeleteOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_DeleteOptions(a.(*url.Values), b.(*DeleteOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*DeleteOptions)(nil), (*DeleteOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_DeleteOptions_To_v1_DeleteOptions(a.(*DeleteOptions), b.(*DeleteOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*Duration)(nil), (**Duration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_Duration_To_Pointer_v1_Duration(a.(*Duration), b.(**Duration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*InternalEvent)(nil), (*WatchEvent)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_InternalEvent_To_v1_WatchEvent(a.(*InternalEvent), b.(*WatchEvent), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*LabelSelector)(nil), (*map[string]string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_LabelSelector_To_Map_string_To_string(a.(*LabelSelector), b.(*map[string]string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*ListMeta)(nil), (*ListMeta)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_ListMeta_To_v1_ListMeta(a.(*ListMeta), b.(*ListMeta), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*MicroTime)(nil), (*MicroTime)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_MicroTime_To_v1_MicroTime(a.(*MicroTime), b.(*MicroTime), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*Time)(nil), (*Time)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_Time_To_v1_Time(a.(*Time), b.(*Time), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*TypeMeta)(nil), (*TypeMeta)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_TypeMeta_To_v1_TypeMeta(a.(*TypeMeta), b.(*TypeMeta), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*WatchEvent)(nil), (*InternalEvent)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_WatchEvent_To_v1_InternalEvent(a.(*WatchEvent), b.(*InternalEvent), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*WatchEvent)(nil), (*watch.Event)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_WatchEvent_To_watch_Event(a.(*WatchEvent), b.(*watch.Event), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*watch.Event)(nil), (*WatchEvent)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_watch_Event_To_v1_WatchEvent(a.(*watch.Event), b.(*WatchEvent), scope)
	}); err != nil {
		return err
	}
	return nil
}

// p282
func autoConvert_url_Values_To_v1_CreateOptions(in *url.Values, out *CreateOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["fieldManager"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldManager, s); err != nil {
			return err
		}
	} else {
		out.FieldManager = ""
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

// Convert_url_Values_To_v1_CreateOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_CreateOptions(in *url.Values, out *CreateOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_CreateOptions(in, out, s)
}

// p312
func autoConvert_url_Values_To_v1_DeleteOptions(in *url.Values, out *DeleteOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["gracePeriodSeconds"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.GracePeriodSeconds, s); err != nil {
			return err
		}
	} else {
		out.GracePeriodSeconds = nil
	}

	if values, ok := map[string][]string(*in)["orphanDependents"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_bool(&values, &out.OrphanDependents, s); err != nil {
			return err
		}
	} else {
		out.OrphanDependents = nil
	}

	if values, ok := map[string][]string(*in)["propagationPolicy"]; ok && len(values) > 0 {
		if err := Convert_Slice_string_To_Pointer_v1_DeletionPropagation(&values, &out.PropagationPolicy, s); err != nil {
			return err
		}
	} else {
		out.PropagationPolicy = nil
	}

	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	return nil
}

func autoConvert_url_Values_To_v1_GetOptions(in *url.Values, out *GetOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["resourceVersion"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.ResourceVersion, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersion = ""
	}
	return nil
}

// Convert_url_Values_To_v1_GetOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_GetOptions(in *url.Values, out *GetOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_GetOptions(in, out, s)
}

func autoConvert_url_Values_To_v1_ListOptions(in *url.Values, out *ListOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["labelSelector"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.LabelSelector, s); err != nil {
			return err
		}
	} else {
		out.LabelSelector = ""
	}
	if values, ok := map[string][]string(*in)["fieldSelector"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldSelector, s); err != nil {
			return err
		}
	} else {
		out.FieldSelector = ""
	}
	if values, ok := map[string][]string(*in)["watch"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.Watch, s); err != nil {
			return err
		}
	} else {
		out.Watch = false
	}
	if values, ok := map[string][]string(*in)["allowWatchBookmarks"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.AllowWatchBookmarks, s); err != nil {
			return err
		}
	} else {
		out.AllowWatchBookmarks = false
	}
	if values, ok := map[string][]string(*in)["resourceVersion"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.ResourceVersion, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersion = ""
	}
	if values, ok := map[string][]string(*in)["resourceVersionMatch"]; ok && len(values) > 0 {
		if err := Convert_Slice_string_To_v1_ResourceVersionMatch(&values, &out.ResourceVersionMatch, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersionMatch = ""
	}
	if values, ok := map[string][]string(*in)["timeoutSeconds"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.TimeoutSeconds, s); err != nil {
			return err
		}
	} else {
		out.TimeoutSeconds = nil
	}
	if values, ok := map[string][]string(*in)["limit"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_int64(&values, &out.Limit, s); err != nil {
			return err
		}
	} else {
		out.Limit = 0
	}
	if values, ok := map[string][]string(*in)["continue"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.Continue, s); err != nil {
			return err
		}
	} else {
		out.Continue = ""
	}
	return nil
}

// Convert_url_Values_To_v1_ListOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_ListOptions(in *url.Values, out *ListOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_ListOptions(in, out, s)
}

func autoConvert_url_Values_To_v1_PatchOptions(in *url.Values, out *PatchOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["force"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_bool(&values, &out.Force, s); err != nil {
			return err
		}
	} else {
		out.Force = nil
	}
	if values, ok := map[string][]string(*in)["fieldManager"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldManager, s); err != nil {
			return err
		}
	} else {
		out.FieldManager = ""
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

// Convert_url_Values_To_v1_PatchOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_PatchOptions(in *url.Values, out *PatchOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_PatchOptions(in, out, s)
}

func autoConvert_url_Values_To_v1_TableOptions(in *url.Values, out *TableOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["-"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.NoHeaders, s); err != nil {
			return err
		}
	} else {
		out.NoHeaders = false
	}
	if values, ok := map[string][]string(*in)["includeObject"]; ok && len(values) > 0 {
		if err := Convert_Slice_string_To_v1_IncludeObjectPolicy(&values, &out.IncludeObject, s); err != nil {
			return err
		}
	} else {
		out.IncludeObject = ""
	}
	return nil
}

// Convert_url_Values_To_v1_TableOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_TableOptions(in *url.Values, out *TableOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_TableOptions(in, out, s)
}

func autoConvert_url_Values_To_v1_UpdateOptions(in *url.Values, out *UpdateOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.

	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["fieldManager"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldManager, s); err != nil {
			return err
		}
	} else {
		out.FieldManager = ""
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

// Convert_url_Values_To_v1_UpdateOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_UpdateOptions(in *url.Values, out *UpdateOptions, s conversion.Scope) error {
	return autoConvert_url_Values_To_v1_UpdateOptions(in, out, s)
}
