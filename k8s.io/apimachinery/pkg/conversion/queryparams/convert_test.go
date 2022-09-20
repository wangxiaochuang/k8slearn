package queryparams_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion/queryparams"
)

type childStructs struct {
	Container      string       `json:"container,omitempty"`
	Follow         bool         `json:"follow,omitempty"`
	Previous       bool         `json:"previous,omitempty"`
	SinceSeconds   *int64       `json:"sinceSeconds,omitempty"`
	TailLines      *int64       `json:"tailLines,omitempty"`
	SinceTime      *metav1.Time `json:"sinceTime,omitempty"`
	EmptyTime      *metav1.Time `json:"emptyTime"`
	NonPointerTime metav1.Time  `json:"nonPointerTime"`
}

func TestConvert(t *testing.T) {
	sinceSeconds := int64(123)
	tailLines := int64(0)
	sinceTime := metav1.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)
	input := &childStructs{
		Container:      "mycontainer",
		Follow:         true,
		Previous:       true,
		SinceSeconds:   &sinceSeconds,
		TailLines:      &tailLines,
		SinceTime:      nil, // test a nil custom marshaller with omitempty
		NonPointerTime: sinceTime,
	}
	queryparams.Convert(input)
}
