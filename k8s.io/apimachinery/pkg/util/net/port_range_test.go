package net

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestPortRange(t *testing.T) {
	pr := &PortRange{}
	var f pflag.Value = pr
	err := f.Set("8000-8888")
	if err != nil {
		t.Error(err)
	}
	r, err := ParsePortRange("8000+2000")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v", r)
}
