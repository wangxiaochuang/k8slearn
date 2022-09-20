package intstr

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

type IntOrString struct {
	Type   Type   `protobuf:"varint,1,opt,name=type,casttype=Type"`
	IntVal int32  `protobuf:"varint,2,opt,name=intVal"`
	StrVal string `protobuf:"bytes,3,opt,name=strVal"`
}

type Type int64

const (
	Int    Type = iota // The IntOrString holds an int.
	String             // The IntOrString holds a string.
)

func FromInt(val int) IntOrString {
	if val > math.MaxInt32 || val < math.MinInt32 {
		klog.Errorf("value: %d overflows int32\n%s\n", val, debug.Stack())
	}
	return IntOrString{Type: Int, IntVal: int32(val)}
}

func FromString(val string) IntOrString {
	return IntOrString{Type: String, StrVal: val}
}

func Parse(val string) IntOrString {
	i, err := strconv.Atoi(val)
	if err != nil {
		return FromString(val)
	}
	return FromInt(i)
}

func (intstr *IntOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		intstr.Type = String
		return json.Unmarshal(value, &intstr.StrVal)
	}
	intstr.Type = Int
	return json.Unmarshal(value, &intstr.IntVal)
}

func (intstr *IntOrString) String() string {
	if intstr == nil {
		return "<nil>"
	}
	if intstr.Type == String {
		return intstr.StrVal
	}
	return strconv.Itoa(intstr.IntValue())
}

func (intstr *IntOrString) IntValue() int {
	if intstr.Type == String {
		i, _ := strconv.Atoi(intstr.StrVal)
		return i
	}
	return int(intstr.IntVal)
}

func (intstr IntOrString) MarshalJSON() ([]byte, error) {
	switch intstr.Type {
	case Int:
		return json.Marshal(intstr.IntVal)
	case String:
		return json.Marshal(intstr.StrVal)
	default:
		return []byte{}, fmt.Errorf("impossible IntOrString.Type")
	}
}

func (IntOrString) OpenAPISchemaType() []string { return []string{"string"} }

func (IntOrString) OpenAPISchemaFormat() string { return "int-or-string" }

func (IntOrString) OpenAPIV3OneOfTypes() []string { return []string{"integer", "string"} }

func ValueOrDefault(intOrPercent *IntOrString, defaultValue IntOrString) *IntOrString {
	if intOrPercent == nil {
		return &defaultValue
	}
	return intOrPercent
}

func GetScaledValueFromIntOrPercent(intOrPercent *IntOrString, total int, roundUp bool) (int, error) {
	if intOrPercent == nil {
		return 0, errors.New("nil value for IntOrString")
	}
	value, isPercent, err := getIntOrPercentValueSafely(intOrPercent)
	if err != nil {
		return 0, fmt.Errorf("invalid value for IntOrString: %v", err)
	}
	if isPercent {
		if roundUp {
			value = int(math.Ceil(float64(value) * (float64(total)) / 100))
		} else {
			value = int(math.Floor(float64(value) * (float64(total)) / 100))
		}
	}
	return value, nil
}

func GetValueFromIntOrPercent(intOrPercent *IntOrString, total int, roundUp bool) (int, error) {
	if intOrPercent == nil {
		return 0, errors.New("nil value for IntOrString")
	}
	value, isPercent, err := getIntOrPercentValue(intOrPercent)
	if err != nil {
		return 0, fmt.Errorf("invalid value for IntOrString: %v", err)
	}
	if isPercent {
		if roundUp {
			value = int(math.Ceil(float64(value) * (float64(total)) / 100))
		} else {
			value = int(math.Floor(float64(value) * (float64(total)) / 100))
		}
	}
	return value, nil
}

func getIntOrPercentValue(intOrStr *IntOrString) (int, bool, error) {
	switch intOrStr.Type {
	case Int:
		return intOrStr.IntValue(), false, nil
	case String:
		s := strings.Replace(intOrStr.StrVal, "%", "", -1)
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0, false, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(v), true, nil
	}
	return 0, false, fmt.Errorf("invalid type: neither int nor percentage")
}

func getIntOrPercentValueSafely(intOrStr *IntOrString) (int, bool, error) {
	switch intOrStr.Type {
	case Int:
		return intOrStr.IntValue(), false, nil
	case String:
		isPercent := false
		s := intOrStr.StrVal
		if strings.HasSuffix(s, "%") {
			isPercent = true
			s = strings.TrimSuffix(intOrStr.StrVal, "%")
		} else {
			return 0, false, fmt.Errorf("invalid type: string is not a percentage")
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0, false, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(v), isPercent, nil
	}
	return 0, false, fmt.Errorf("invalid type: neither int nor percentage")
}
