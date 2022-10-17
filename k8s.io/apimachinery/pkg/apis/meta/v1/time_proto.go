package v1

import "time"

type Timestamp struct {
	Seconds int64 `json:"seconds" protobuf:"varint,1,opt,name=seconds"`
	Nanos   int32 `json:"nanos" protobuf:"varint,2,opt,name=nanos"`
}

func (m *Time) ProtoTime() *Timestamp {
	if m == nil {
		return &Timestamp{}
	}
	return &Timestamp{
		Seconds: m.Time.Unix(),
	}
}

func (m *Time) Size() (n int) {
	if m == nil || m.Time.IsZero() {
		return 0
	}
	return m.ProtoTime().Size()
}

func (m *Time) Unmarshal(data []byte) error {
	if len(data) == 0 {
		m.Time = time.Time{}
		return nil
	}
	p := Timestamp{}
	if err := p.Unmarshal(data); err != nil {
		return err
	}
	m.Time = time.Unix(p.Seconds, int64(0)).Local()
	return nil
}

func (m *Time) Marshal() (data []byte, err error) {
	if m == nil || m.Time.IsZero() {
		return nil, nil
	}
	return m.ProtoTime().Marshal()
}

func (m *Time) MarshalTo(data []byte) (int, error) {
	if m == nil || m.Time.IsZero() {
		return 0, nil
	}
	return m.ProtoTime().MarshalTo(data)
}

func (m *Time) MarshalToSizedBuffer(data []byte) (int, error) {
	if m == nil || m.Time.IsZero() {
		return 0, nil
	}
	return m.ProtoTime().MarshalToSizedBuffer(data)
}
