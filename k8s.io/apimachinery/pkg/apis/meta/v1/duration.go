package v1

import (
	"encoding/json"
	"time"
)

type Duration struct {
	time.Duration `protobuf:"varint,1,opt,name=duration,casttype=time.Duration"`
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pd, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	d.Duration = pd
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

func (d Duration) ToUnstructured() interface{} {
	return d.Duration.String()
}

func (_ Duration) OpenAPISchemaType() []string { return []string{"string"} }

func (_ Duration) OpenAPISchemaFormat() string { return "" }
