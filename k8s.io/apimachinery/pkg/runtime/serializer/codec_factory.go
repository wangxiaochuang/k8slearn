package serializer

import (
	"mime"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/apimachinery/pkg/runtime/serializer/recognizer"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
)

// serializerExtensions are for serializers that are conditionally compiled in
var serializerExtensions = []func(*runtime.Scheme) (serializerType, bool){}

type serializerType struct {
	AcceptContentTypes []string
	ContentType        string
	FileExtensions     []string
	// EncodesAsText should be true if this content type can be represented safely in UTF-8
	EncodesAsText bool

	// Encoder Decoder
	Serializer       runtime.Serializer
	PrettySerializer runtime.Serializer
	StrictSerializer runtime.Serializer

	AcceptStreamContentTypes []string
	StreamContentType        string

	// NewFrameReader NewFrameWriter
	Framer runtime.Framer
	// Encoder Decoder
	StreamSerializer runtime.Serializer
}

// p52
func newSerializersForScheme(scheme *runtime.Scheme, mf json.MetaFactory, options CodecFactoryOptions) []serializerType {
	jsonSerializer := json.NewSerializerWithOptions(
		mf, scheme, scheme,
		json.SerializerOptions{Yaml: false, Pretty: false, Strict: options.Strict},
	)
	jsonSerializerType := serializerType{
		AcceptContentTypes: []string{runtime.ContentTypeJSON},
		ContentType:        runtime.ContentTypeJSON,
		FileExtensions:     []string{"json"},
		EncodesAsText:      true,
		Serializer:         jsonSerializer,

		Framer:           json.Framer,
		StreamSerializer: jsonSerializer,
	}
	// 放后面是因为StreamSerializer是不需要pretty的
	if options.Pretty {
		jsonSerializerType.PrettySerializer = json.NewSerializerWithOptions(
			mf, scheme, scheme,
			json.SerializerOptions{Yaml: false, Pretty: true, Strict: options.Strict},
		)
	}

	strictJSONSerializer := json.NewSerializerWithOptions(
		mf, scheme, scheme,
		json.SerializerOptions{Yaml: false, Pretty: false, Strict: true},
	)
	jsonSerializerType.StrictSerializer = strictJSONSerializer

	yamlSerializer := json.NewSerializerWithOptions(
		mf, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: true},
	)
	strictYAMLSerializer := json.NewSerializerWithOptions(
		mf, scheme, scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: true},
	)
	protoSerializer := protobuf.NewSerializer(scheme, scheme)
	protoRawSerializer := protobuf.NewRawSerializer(scheme, scheme)

	serializers := []serializerType{
		jsonSerializerType,
		{
			AcceptContentTypes: []string{runtime.ContentTypeYAML},
			ContentType:        runtime.ContentTypeYAML,
			FileExtensions:     []string{"yaml"},
			EncodesAsText:      true,
			Serializer:         yamlSerializer,
			StrictSerializer:   strictYAMLSerializer,
		},
		{
			AcceptContentTypes: []string{runtime.ContentTypeProtobuf},
			ContentType:        runtime.ContentTypeProtobuf,
			FileExtensions:     []string{"pb"},
			Serializer:         protoSerializer,
			// note, strict decoding is unsupported for protobuf,
			// fall back to regular serializing
			StrictSerializer: protoSerializer,

			Framer:           protobuf.LengthDelimitedFramer,
			StreamSerializer: protoRawSerializer,
		},
	}
	for _, fn := range serializerExtensions {
		if serializer, ok := fn(scheme); ok {
			serializers = append(serializers, serializer)
		}
	}
	return serializers
}

// CodecFactory provides methods for retrieving codecs and serializers for specific
// versions and content types.
type CodecFactory struct {
	scheme    *runtime.Scheme
	universal runtime.Decoder
	accepts   []runtime.SerializerInfo

	legacySerializer runtime.Serializer
}

// p134
type CodecFactoryOptions struct {
	Strict bool
	Pretty bool
}

// p143
type CodecFactoryOptionsMutator func(*CodecFactoryOptions)

// EnablePretty enables including a pretty serializer along with the non-pretty one
func EnablePretty(options *CodecFactoryOptions) {
	options.Pretty = true
}

// DisablePretty disables including a pretty serializer along with the non-pretty one
func DisablePretty(options *CodecFactoryOptions) {
	options.Pretty = false
}

// EnableStrict enables configuring all serializers in strict mode
func EnableStrict(options *CodecFactoryOptions) {
	options.Strict = true
}

// DisableStrict disables configuring all serializers in strict mode
func DisableStrict(options *CodecFactoryOptions) {
	options.Strict = false
}

// p176 mutators就是用来自定义options参数的
func NewCodecFactory(scheme *runtime.Scheme, mutators ...CodecFactoryOptionsMutator) CodecFactory {
	options := CodecFactoryOptions{Pretty: true}
	for _, fn := range mutators {
		fn(&options)
	}

	// MetaFactory就是尝试从byte数组中解析gvk
	// 返回的是一个解析器的数组，支持yaml、json、protobuf
	serializers := newSerializersForScheme(scheme, json.DefaultMetaFactory, options)
	return newCodecFactory(scheme, serializers)
}

// newCodecFactory is a helper for testing that allows a different metafactory to be specified.
func newCodecFactory(scheme *runtime.Scheme, serializers []serializerType) CodecFactory {
	decoders := make([]runtime.Decoder, 0, len(serializers))
	var accepts []runtime.SerializerInfo
	alreadyAccepted := make(map[string]struct{})

	var legacySerializer runtime.Serializer
	for _, d := range serializers {
		// 把所有的序列化器都放到一个数组
		decoders = append(decoders, d.Serializer)
		// json的就是application/json yaml的就是application/yaml
		for _, mediaType := range d.AcceptContentTypes {
			if _, ok := alreadyAccepted[mediaType]; ok {
				continue
			}
			// 把所有接收的mediaType都放map里
			alreadyAccepted[mediaType] = struct{}{}
			// 转成runtime.SerializerInfo
			info := runtime.SerializerInfo{
				MediaType:        d.ContentType,
				EncodesAsText:    d.EncodesAsText,
				Serializer:       d.Serializer,
				PrettySerializer: d.PrettySerializer,
				StrictSerializer: d.StrictSerializer,
			}
			mediaType, _, err := mime.ParseMediaType(info.MediaType)
			if err != nil {
				panic(err)
			}
			parts := strings.SplitN(mediaType, "/", 2)
			// application json
			info.MediaTypeType = parts[0]
			info.MediaTypeSubType = parts[1]

			if d.StreamSerializer != nil {
				info.StreamSerializer = &runtime.StreamSerializerInfo{
					Serializer:    d.StreamSerializer,
					EncodesAsText: d.EncodesAsText,
					Framer:        d.Framer,
				}
			}
			// runtime.SerializerInfo的一个数组
			accepts = append(accepts, info)
			if mediaType == runtime.ContentTypeJSON {
				legacySerializer = d.Serializer
			}
		}
	}
	if legacySerializer == nil {
		legacySerializer = serializers[0].Serializer
	}
	return CodecFactory{
		scheme: scheme,
		// 这里汇集了多个编解码器，会依次尝试，yaml会先转为json
		universal:        recognizer.NewDecoder(decoders...),
		accepts:          accepts,
		legacySerializer: legacySerializer,
	}
}

// WithoutConversion returns a NegotiatedSerializer that performs no conversion, even if the
// caller requests it.
func (f CodecFactory) WithoutConversion() runtime.NegotiatedSerializer {
	return WithoutConversionCodecFactory{f}
}

// SupportedMediaTypes returns the RFC2046 media types that this factory has serializers for.
func (f CodecFactory) SupportedMediaTypes() []runtime.SerializerInfo {
	return f.accepts
}

// LegacyCodec encodes output to a given API versions, and decodes output into the internal form from
// any recognized source. The returned codec will always encode output to JSON. If a type is not
// found in the list of versions an error will be returned.
//
// This method is deprecated - clients and servers should negotiate a serializer by mime-type and
// invoke CodecForVersions. Callers that need only to read data should use UniversalDecoder().
//
// TODO: make this call exist only in pkg/api, and initialize it with the set of default versions.
//
//	All other callers will be forced to request a Codec directly.
func (f CodecFactory) LegacyCodec(version ...schema.GroupVersion) runtime.Codec {
	return versioning.NewDefaultingCodecForScheme(f.scheme, f.legacySerializer, f.universal, schema.GroupVersions(version), runtime.InternalGroupVersioner)
}

// UniversalDeserializer can convert any stored data recognized by this factory into a Go object that satisfies
// runtime.Object. It does not perform conversion. It does not perform defaulting.
func (f CodecFactory) UniversalDeserializer() runtime.Decoder {
	return f.universal
}

// 返回runtime.Decoder，可以指定版本，默认按内部版本decode
func (f CodecFactory) UniversalDecoder(versions ...schema.GroupVersion) runtime.Decoder {
	var versioner runtime.GroupVersioner
	if len(versions) == 0 {
		versioner = runtime.InternalGroupVersioner
	} else {
		versioner = schema.GroupVersions(versions)
	}
	// 只专递了Decode的参数和期望的版本
	return f.CodecForVersions(nil, f.universal, nil, versioner)
}

// 使用提供的序列化器创建一个编解码器，没有指定就是内部版本
func (f CodecFactory) CodecForVersions(encoder runtime.Encoder, decoder runtime.Decoder, encode runtime.GroupVersioner, decode runtime.GroupVersioner) runtime.Codec {
	// TODO: these are for backcompat, remove them in the future
	if encode == nil {
		// encode没有指定，就不解析
		encode = runtime.DisabledGroupVersioner
	}
	if decode == nil {
		// decode没指定，就默认解析为内部版本
		decode = runtime.InternalGroupVersioner
	}
	return versioning.NewDefaultingCodecForScheme(f.scheme, encoder, decoder, encode, decode)
}

// DecoderToVersion returns a decoder that targets the provided group version.
func (f CodecFactory) DecoderToVersion(decoder runtime.Decoder, gv runtime.GroupVersioner) runtime.Decoder {
	return f.CodecForVersions(nil, decoder, nil, gv)
}

// EncoderForVersion returns an encoder that targets the provided group version.
func (f CodecFactory) EncoderForVersion(encoder runtime.Encoder, gv runtime.GroupVersioner) runtime.Encoder {
	return f.CodecForVersions(encoder, nil, gv, nil)
}

// WithoutConversionCodecFactory is a CodecFactory that will explicitly ignore requests to perform conversion.
// This wrapper is used while code migrates away from using conversion (such as external clients) and in the future
// will be unnecessary when we change the signature of NegotiatedSerializer.
type WithoutConversionCodecFactory struct {
	CodecFactory
}

// EncoderForVersion returns an encoder that does not do conversion, but does set the group version kind of the object
// when serialized.
func (f WithoutConversionCodecFactory) EncoderForVersion(serializer runtime.Encoder, version runtime.GroupVersioner) runtime.Encoder {
	return runtime.WithVersionEncoder{
		Version:     version,
		Encoder:     serializer,
		ObjectTyper: f.CodecFactory.scheme,
	}
}

// DecoderToVersion returns an decoder that does not do conversion.
func (f WithoutConversionCodecFactory) DecoderToVersion(serializer runtime.Decoder, _ runtime.GroupVersioner) runtime.Decoder {
	return runtime.WithoutVersionDecoder{
		Decoder: serializer,
	}
}
