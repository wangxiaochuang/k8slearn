package runtime

type TypeMeta struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty" protobuf:"bytes,1,opt,name=apiVersion"`
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty" protobuf:"bytes,2,opt,name=kind"`
}

const (
	ContentTypeJSON     string = "application/json"
	ContentTypeYAML     string = "application/yaml"
	ContentTypeProtobuf string = "application/vnd.kubernetes.protobuf"
)

type RawExtension struct {
	Raw    []byte `json:"-" protobuf:"bytes,1,opt,name=raw"`
	Object Object `json:"-"`
}

type Unknown struct {
	TypeMeta        `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	Raw             []byte `protobuf:"bytes,2,opt,name=raw"`
	ContentEncoding string `protobuf:"bytes,3,opt,name=contentEncoding"`
	ContentType     string `protobuf:"bytes,4,opt,name=contentType"`
}
