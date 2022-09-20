package v1

import "k8s.io/apimachinery/pkg/runtime"

// p492
func (in *LabelSelector) DeepCopyInto(out *LabelSelector) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.MatchExpressions != nil {
		in, out := &in.MatchExpressions, &out.MatchExpressions
		*out = make([]LabelSelectorRequirement, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

func (in *LabelSelector) DeepCopy() *LabelSelector {
	if in == nil {
		return nil
	}
	out := new(LabelSelector)
	in.DeepCopyInto(out)
	return out
}

func (in *LabelSelectorRequirement) DeepCopyInto(out *LabelSelectorRequirement) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *LabelSelectorRequirement) DeepCopy() *LabelSelectorRequirement {
	if in == nil {
		return nil
	}
	out := new(LabelSelectorRequirement)
	in.DeepCopyInto(out)
	return out
}

// p576
func (in *ListMeta) DeepCopyInto(out *ListMeta) {
	*out = *in
	if in.RemainingItemCount != nil {
		in, out := &in.RemainingItemCount, &out.RemainingItemCount
		*out = new(int64)
		**out = **in
	}
	return
}

func (in *ListMeta) DeepCopy() *ListMeta {
	if in == nil {
		return nil
	}
	out := new(ListMeta)
	in.DeepCopyInto(out)
	return out
}

// p920
func (in *Status) DeepCopyInto(out *Status) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Details != nil {
		in, out := &in.Details, &out.Details
		*out = new(StatusDetails)
		(*in).DeepCopyInto(*out)
	}
	return
}

func (in *Status) DeepCopy() *Status {
	if in == nil {
		return nil
	}
	out := new(Status)
	in.DeepCopyInto(out)
	return out
}

func (in *Status) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// p967
func (in *StatusDetails) DeepCopyInto(out *StatusDetails) {
	*out = *in
	if in.Causes != nil {
		in, out := &in.Causes, &out.Causes
		*out = make([]StatusCause, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *StatusDetails) DeepCopy() *StatusDetails {
	if in == nil {
		return nil
	}
	out := new(StatusDetails)
	in.DeepCopyInto(out)
	return out
}
