package flag

import (
	"fmt"
	"sort"
	"strings"
)

type MapStringString struct {
	Map         *map[string]string
	initialized bool
	NoSplit     bool
}

func NewMapStringString(m *map[string]string) *MapStringString {
	return &MapStringString{Map: m}
}

func NewMapStringStringNoSplit(m *map[string]string) *MapStringString {
	return &MapStringString{
		Map:     m,
		NoSplit: true,
	}
}

func (m *MapStringString) String() string {
	if m == nil || m.Map == nil {
		return ""
	}
	pairs := []string{}
	for k, v := range *m.Map {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func (m *MapStringString) Set(value string) error {
	if m.Map == nil {
		return fmt.Errorf("no target (nil pointer to map[string]string)")
	}
	if !m.initialized || *m.Map == nil {
		// clear default values, or allocate if no existing map
		*m.Map = make(map[string]string)
		m.initialized = true
	}

	// account for comma-separated key-value pairs in a single invocation
	if !m.NoSplit {
		for _, s := range strings.Split(value, ",") {
			if len(s) == 0 {
				continue
			}
			arr := strings.SplitN(s, "=", 2)
			if len(arr) != 2 {
				return fmt.Errorf("malformed pair, expect string=string")
			}
			k := strings.TrimSpace(arr[0])
			v := strings.TrimSpace(arr[1])
			(*m.Map)[k] = v
		}
		return nil
	}

	// account for only one key-value pair in a single invocation
	arr := strings.SplitN(value, "=", 2)
	if len(arr) != 2 {
		return fmt.Errorf("malformed pair, expect string=string")
	}
	k := strings.TrimSpace(arr[0])
	v := strings.TrimSpace(arr[1])
	(*m.Map)[k] = v
	return nil

}

func (*MapStringString) Type() string {
	return "mapStringString"
}

// Empty implements OmitEmpty
func (m *MapStringString) Empty() bool {
	return len(*m.Map) == 0
}
