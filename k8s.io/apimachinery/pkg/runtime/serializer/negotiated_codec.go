package serializer

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// TODO: We should split negotiated serializers that we can change versions on from those we can change
// serialization formats on
type negotiatedSerializerWrapper struct {
	info runtime.SerializerInfo
}

func NegotiatedSerializerWrapper(info runtime.SerializerInfo) runtime.NegotiatedSerializer {
	return &negotiatedSerializerWrapper{info}
}

func (n *negotiatedSerializerWrapper) SupportedMediaTypes() []runtime.SerializerInfo {
	return []runtime.SerializerInfo{n.info}
}

func (n *negotiatedSerializerWrapper) EncoderForVersion(e runtime.Encoder, _ runtime.GroupVersioner) runtime.Encoder {
	return e
}

func (n *negotiatedSerializerWrapper) DecoderToVersion(d runtime.Decoder, _gv runtime.GroupVersioner) runtime.Decoder {
	return d
}
