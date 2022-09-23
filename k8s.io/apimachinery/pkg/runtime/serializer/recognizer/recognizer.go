package recognizer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RecognizingDecoder interface {
	runtime.Decoder
	RecognizesData(peek []byte) (ok, unknown bool, err error)
}

// NewDecoder creates a decoder that will attempt multiple decoders in an order defined
// by:
//
// 1. The decoder implements RecognizingDecoder and identifies the data
// 2. All other decoders, and any decoder that returned true for unknown.
//
// The order passed to the constructor is preserved within those priorities.
func NewDecoder(decoders ...runtime.Decoder) runtime.Decoder {
	return &decoder{
		decoders: decoders,
	}
}

type decoder struct {
	decoders []runtime.Decoder
}

var _ RecognizingDecoder = &decoder{}

func (d *decoder) RecognizesData(data []byte) (bool, bool, error) {
	var (
		lastErr    error
		anyUnknown bool
	)
	for _, r := range d.decoders {
		switch t := r.(type) {
		case RecognizingDecoder:
			ok, unknown, err := t.RecognizesData(data)
			if err != nil {
				lastErr = err
				continue
			}
			anyUnknown = anyUnknown || unknown
			if !ok {
				continue
			}
			return true, false, nil
		}
	}
	return false, anyUnknown, lastErr
}

func (d *decoder) Decode(data []byte, gvk *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	var (
		lastErr error
		skipped []runtime.Decoder
	)

	// 尝试使用每个编码器进行解码
	for _, r := range d.decoders {
		switch t := r.(type) {
		case RecognizingDecoder:
			ok, unknown, err := t.RecognizesData(data)
			if err != nil {
				lastErr = err
				continue
			}
			// 如果是未知的就放数组里，后面可能都去尝试一下
			if unknown {
				skipped = append(skipped, t)
				continue
			}
			// 明确不知道的，直接继续下一个
			if !ok {
				continue
			}
			// 这是识别到了是什么类型，直接Decode即可
			return r.Decode(data, gvk, into)
		default:
			skipped = append(skipped, t)
		}
	}

	// 尝试每一个不确定的解码器
	for _, r := range skipped {
		out, actual, err := r.Decode(data, gvk, into)
		if err != nil {
			// if we got an object back from the decoder, and the
			// error was a strict decoding error (e.g. unknown or
			// duplicate fields), we still consider the recognizer
			// to have understood the object
			if out == nil || !runtime.IsStrictDecodingError(err) {
				lastErr = err
				continue
			}
		}
		return out, actual, err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no serialization format matched the provided data")
	}
	return nil, nil, lastErr
}
