package runtime

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NegotiateError struct {
	ContentType string
	Stream      bool
}

func (e NegotiateError) Error() string {
	if e.Stream {
		return fmt.Sprintf("no stream serializers registered for %s", e.ContentType)
	}
	return fmt.Sprintf("no serializers registered for %s", e.ContentType)
}

type clientNegotiator struct {
	serializer     NegotiatedSerializer
	encode, decode GroupVersioner
}

func (n *clientNegotiator) Encoder(contentType string, params map[string]string) (Encoder, error) {
	panic("not implemented")
}

func (n *clientNegotiator) Decoder(contentType string, params map[string]string) (Decoder, error) {
	panic("not implemented")
}

func (n *clientNegotiator) StreamDecoder(contentType string, params map[string]string) (Decoder, Serializer, Framer, error) {
	panic("not implemented")
}

func NewClientNegotiator(serializer NegotiatedSerializer, gv schema.GroupVersion) ClientNegotiator {
	return &clientNegotiator{
		serializer: serializer,
		encode:     gv,
	}
}

type simpleNegotiatedSerializer struct {
	info SerializerInfo
}

func NewSimpleNegotiatedSerializer(info SerializerInfo) NegotiatedSerializer {
	return &simpleNegotiatedSerializer{info: info}
}

func (n *simpleNegotiatedSerializer) SupportedMediaTypes() []SerializerInfo {
	return []SerializerInfo{n.info}
}

func (n *simpleNegotiatedSerializer) EncoderForVersion(e Encoder, _ GroupVersioner) Encoder {
	return e
}

func (n *simpleNegotiatedSerializer) DecoderToVersion(d Decoder, _gv GroupVersioner) Decoder {
	return d
}
