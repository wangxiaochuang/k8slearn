package fuzzer

import (
	"math/rand"

	fuzz "github.com/google/gofuzz"

	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

// FuzzerFuncs returns a list of func(*SomeType, c fuzz.Continue) functions.
type FuzzerFuncs func(codecs runtimeserializer.CodecFactory) []interface{}

// FuzzerFor can randomly populate api objects that are destined for version.
func FuzzerFor(funcs FuzzerFuncs, src rand.Source, codecs runtimeserializer.CodecFactory) *fuzz.Fuzzer {
	f := fuzz.New().NilChance(.5).NumElements(0, 1)
	if src != nil {
		f.RandSource(src)
	}
	f.Funcs(funcs(codecs)...)
	return f
}

// MergeFuzzerFuncs will merge the given funcLists, overriding early funcs with later ones if there first
// argument has the same type.
func MergeFuzzerFuncs(funcs ...FuzzerFuncs) FuzzerFuncs {
	return FuzzerFuncs(func(codecs runtimeserializer.CodecFactory) []interface{} {
		result := []interface{}{}
		for _, f := range funcs {
			if f != nil {
				result = append(result, f(codecs)...)
			}
		}
		return result
	})
}
