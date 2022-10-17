package watch

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/watch"
)

type Decoder struct {
	decoder         streaming.Decoder
	embeddedDecoder runtime.Decoder
}

func NewDecoder(decoder streaming.Decoder, embeddedDecoder runtime.Decoder) *Decoder {
	return &Decoder{
		decoder:         decoder,
		embeddedDecoder: embeddedDecoder,
	}
}

func (d *Decoder) Decode() (watch.EventType, runtime.Object, error) {
	var got metav1.WatchEvent
	res, _, err := d.decoder.Decode(nil, &got)
	if err != nil {
		return "", nil, err
	}
	if res != &got {
		return "", nil, fmt.Errorf("unable to decode to metav1.Event")
	}
	switch got.Type {
	case string(watch.Added), string(watch.Modified), string(watch.Deleted), string(watch.Error), string(watch.Bookmark):
	default:
		return "", nil, fmt.Errorf("got invalid watch event type: %v", got.Type)
	}

	obj, err := runtime.Decode(d.embeddedDecoder, got.Object.Raw)
	if err != nil {
		return "", nil, fmt.Errorf("unable to decode watch event: %v", err)
	}
	return watch.EventType(got.Type), obj, nil
}

// Close closes the underlying r.
func (d *Decoder) Close() {
	d.decoder.Close()
}
