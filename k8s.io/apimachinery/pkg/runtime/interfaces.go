package runtime

import (
	"io"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	APIVersionInternal = "__internal"
)

type GroupVersioner interface {
	KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (target schema.GroupVersionKind, ok bool)
	Identifier() string
}

type Identifier string

type Encoder interface {
	Encode(obj Object, w io.Writer) error
	Identifier() Identifier
}

type MemoryAllocator interface {
	Allocate(n uint64) []byte
}

type EncoderWithAllocator interface {
	Encoder
	EncodeWithAllocator(obj Object, w io.Writer, memAlloc MemoryAllocator) error
}

type Decoder interface {
	Decode(data []byte, defaults *schema.GroupVersionKind, into Object) (Object, *schema.GroupVersionKind, error)
}

type Serializer interface {
	Encoder
	Decoder
}

type Codec Serializer

type ParameterCodec interface {
	DecodeParameters(parameters url.Values, from schema.GroupVersion, into Object) error
	EncodeParameters(obj Object, to schema.GroupVersion) (url.Values, error)
}

type Framer interface {
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	NewFrameWriter(w io.Writer) io.Writer
}

type SerializerInfo struct {
	MediaType        string
	MediaTypeType    string
	MediaTypeSubType string
	EncodesAsText    bool
	Serializer       Serializer
	PrettySerializer Serializer
	StrictSerializer Serializer
	StreamSerializer *StreamSerializerInfo
}

type StreamSerializerInfo struct {
	EncodesAsText bool
	Serializer
	Framer
}

type NegotiatedSerializer interface {
	SupportedMediaTypes() []SerializerInfo
	EncoderForVersion(serializer Encoder, gv GroupVersioner) Encoder
	DecoderToVersion(serializer Decoder, gv GroupVersioner) Decoder
}

type ClientNegotiator interface {
	Encoder(contentType string, params map[string]string) (Encoder, error)
	Decoder(contentType string, params map[string]string) (Decoder, error)
	StreamDecoder(contentType string, params map[string]string) (Decoder, Serializer, Framer, error)
}

type StorageSerializer interface {
	SupportedMediaTypes() []SerializerInfo
	UniversalDeserializer() Decoder
	EncoderForVersion(serializer Encoder, gv GroupVersioner) Encoder
	DecoderToVersion(serializer Decoder, gv GroupVersioner) Decoder
}

type NestedObjectEncoder interface {
	EncodeNestedObjects(e Encoder) error
}

type NestedObjectDecoder interface {
	DecodeNestedObjects(d Decoder) error
}

type ObjectDefaulter interface {
	Default(in Object)
}

type ObjectVersioner interface {
	ConvertToVersion(in Object, gv GroupVersioner) (out Object, err error)
}

// p252
type ObjectConvertor interface {
	Convert(in, out, context interface{}) error
	ConvertToVersion(in Object, gv GroupVersioner) (out Object, err error)
	ConvertFieldLabel(gvk schema.GroupVersionKind, label, value string) (string, string, error)
}

// p269
type ObjectTyper interface {
	ObjectKinds(Object) ([]schema.GroupVersionKind, bool, error)
	Recognizes(gvk schema.GroupVersionKind) bool
}

// p281
type ObjectCreater interface {
	New(kind schema.GroupVersionKind) (out Object, err error)
}

type EquivalentResourceMapper interface {
	EquivalentResourcesFor(resource schema.GroupVersionResource, subresource string) []schema.GroupVersionResource
	KindFor(resource schema.GroupVersionResource, subresource string) schema.GroupVersionKind
}

type EquivalentResourceRegistry interface {
	EquivalentResourceMapper
	RegisterKindFor(resource schema.GroupVersionResource, subresource string, kind schema.GroupVersionKind)
}

type ResourceVersioner interface {
	SetResourceVersion(obj Object, version string) error
	ResourceVersion(obj Object) (string, error)
}

type Namer interface {
	Name(obj Object) (string, error)
	Namespace(obj Object) (string, error)
}

// p323
type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}

type CacheableObject interface {
	CacheEncode(id Identifier, encode func(Object, io.Writer) error, w io.Writer) error
	GetObject() Object
}

type Unstructured interface {
	Object
	NewEmptyInstance() Unstructured

	UnstructuredContent() map[string]interface{}
	SetUnstructuredContent(map[string]interface{})

	IsList() bool

	EachListItem(func(Object) error) error
}
