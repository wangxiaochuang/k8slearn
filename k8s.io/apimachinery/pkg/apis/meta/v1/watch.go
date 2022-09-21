package v1

import (
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

type WatchEvent struct {
	Type   string               `json:"type" protobuf:"bytes,1,opt,name=type"`
	Object runtime.RawExtension `json:"object" protobuf:"bytes,2,opt,name=object"`
}

func Convert_watch_Event_To_v1_WatchEvent(in *watch.Event, out *WatchEvent, s conversion.Scope) error {
	out.Type = string(in.Type)
	switch t := in.Object.(type) {
	case *runtime.Unknown:
		// TODO: handle other fields on Unknown and detect type
		out.Object.Raw = t.Raw
	case nil:
	default:
		out.Object.Object = in.Object
	}
	return nil
}

func Convert_v1_InternalEvent_To_v1_WatchEvent(in *InternalEvent, out *WatchEvent, s conversion.Scope) error {
	return Convert_watch_Event_To_v1_WatchEvent((*watch.Event)(in), out, s)
}

func Convert_v1_WatchEvent_To_watch_Event(in *WatchEvent, out *watch.Event, s conversion.Scope) error {
	out.Type = watch.EventType(in.Type)
	if in.Object.Object != nil {
		out.Object = in.Object.Object
	} else if in.Object.Raw != nil {
		// TODO: handle other fields on Unknown and detect type
		out.Object = &runtime.Unknown{
			Raw:         in.Object.Raw,
			ContentType: runtime.ContentTypeJSON,
		}
	}
	return nil
}

func Convert_v1_WatchEvent_To_v1_InternalEvent(in *WatchEvent, out *InternalEvent, s conversion.Scope) error {
	return Convert_v1_WatchEvent_To_watch_Event(in, (*watch.Event)(out), s)
}

type InternalEvent watch.Event

func (e *InternalEvent) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }
func (e *WatchEvent) GetObjectKind() schema.ObjectKind    { return schema.EmptyObjectKind }
func (e *InternalEvent) DeepCopyObject() runtime.Object {
	if c := e.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}
