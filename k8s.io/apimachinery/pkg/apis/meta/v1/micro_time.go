package v1

import (
	"encoding/json"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"

type MicroTime struct {
	time.Time `protobuf:"-"`
}

func (t *MicroTime) DeepCopyInto(out *MicroTime) {
	*out = *t
}

func NewMicroTime(time time.Time) MicroTime {
	return MicroTime{time}
}

func DateMicro(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) MicroTime {
	return MicroTime{time.Date(year, month, day, hour, min, sec, nsec, loc)}
}

func NowMicro() MicroTime {
	return MicroTime{time.Now()}
}

func (t *MicroTime) IsZero() bool {
	if t == nil {
		return true
	}
	return t.Time.IsZero()
}

func (t *MicroTime) Before(u *MicroTime) bool {
	if t != nil && u != nil {
		return t.Time.Before(u.Time)
	}
	return false
}

func (t *MicroTime) Equal(u *MicroTime) bool {
	if t == nil && u == nil {
		return true
	}
	if t != nil && u != nil {
		return t.Time.Equal(u.Time)
	}
	return false
}

func (t *MicroTime) BeforeTime(u *Time) bool {
	if t != nil && u != nil {
		return t.Time.Before(u.Time)
	}
	return false
}

func (t *MicroTime) EqualTime(u *Time) bool {
	if t == nil && u == nil {
		return true
	}
	if t != nil && u != nil {
		return t.Time.Equal(u.Time)
	}
	return false
}

func UnixMicro(sec int64, nsec int64) MicroTime {
	return MicroTime{time.Unix(sec, nsec)}
}

func (t *MicroTime) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pt, err := time.Parse(RFC3339Micro, str)
	if err != nil {
		return err
	}

	t.Time = pt.Local()
	return nil
}

func (t *MicroTime) UnmarshalQueryParameter(str string) error {
	if len(str) == 0 {
		t.Time = time.Time{}
		return nil
	}
	// Tolerate requests from older clients that used JSON serialization to build query params
	if len(str) == 4 && str == "null" {
		t.Time = time.Time{}
		return nil
	}

	pt, err := time.Parse(RFC3339Micro, str)
	if err != nil {
		return err
	}

	t.Time = pt.Local()
	return nil
}

func (t MicroTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}

	return json.Marshal(t.UTC().Format(RFC3339Micro))
}

func (_ MicroTime) OpenAPISchemaType() []string { return []string{"string"} }

func (_ MicroTime) OpenAPISchemaFormat() string { return "date-time" }

func (t MicroTime) MarshalQueryParameter() (string, error) {
	if t.IsZero() {
		// Encode unset/nil objects as an empty string
		return "", nil
	}

	return t.UTC().Format(RFC3339Micro), nil
}
