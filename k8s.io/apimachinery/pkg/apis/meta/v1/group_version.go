package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GroupResource struct {
	Group    string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Resource string `json:"resource" protobuf:"bytes,2,opt,name=resource"`
}

func (gr *GroupResource) String() string {
	if gr == nil {
		return "<nil>"
	}
	if len(gr.Group) == 0 {
		return gr.Resource
	}
	return gr.Resource + "." + gr.Group
}

type GroupVersionResource struct {
	Group    string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version  string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Resource string `json:"resource" protobuf:"bytes,3,opt,name=resource"`
}

func (gvr *GroupVersionResource) String() string {
	if gvr == nil {
		return "<nil>"
	}
	return strings.Join([]string{gvr.Group, "/", gvr.Version, ", Resource=", gvr.Resource}, "")
}

type GroupKind struct {
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Kind  string `json:"kind" protobuf:"bytes,2,opt,name=kind"`
}

func (gk *GroupKind) String() string {
	if gk == nil {
		return "<nil>"
	}
	if len(gk.Group) == 0 {
		return gk.Kind
	}
	return gk.Kind + "." + gk.Group
}

type GroupVersionKind struct {
	Group   string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Kind    string `json:"kind" protobuf:"bytes,3,opt,name=kind"`
}

func (gvk GroupVersionKind) String() string {
	return gvk.Group + "/" + gvk.Version + ", Kind=" + gvk.Kind
}

type GroupVersion struct {
	Group   string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
}

func (gv GroupVersion) Empty() bool {
	return len(gv.Group) == 0 && len(gv.Version) == 0
}

func (gv GroupVersion) String() string {
	if gv.Empty() {
		return ""
	}

	if len(gv.Group) == 0 && gv.Version == "v1" {
		return gv.Version
	}
	if len(gv.Group) > 0 {
		return gv.Group + "/" + gv.Version
	}
	return gv.Version
}

func (gv GroupVersion) MarshalJSON() ([]byte, error) {
	s := gv.String()
	if strings.Count(s, "/") > 1 {
		return []byte{}, fmt.Errorf("illegal GroupVersion %v: contains more than one /", s)
	}
	return json.Marshal(s)
}

func (gv *GroupVersion) unmarshal(value []byte) error {
	var s string
	if err := json.Unmarshal(value, &s); err != nil {
		return err
	}
	parsed, err := schema.ParseGroupVersion(s)
	if err != nil {
		return err
	}
	gv.Group, gv.Version = parsed.Group, parsed.Version
	return nil
}

func (gv *GroupVersion) UnmarshalJSON(value []byte) error {
	return gv.unmarshal(value)
}

func (gv *GroupVersion) UnmarshalText(value []byte) error {
	return gv.unmarshal(value)
}
