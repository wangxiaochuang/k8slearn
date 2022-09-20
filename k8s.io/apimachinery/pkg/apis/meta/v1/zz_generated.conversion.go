package v1

import (
	"net/url"
	"unsafe"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

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
