package v1

// p107
type UserInfo struct {
	Username string                `json:"username,omitempty" protobuf:"bytes,1,opt,name=username"`
	UID      string                `json:"uid,omitempty" protobuf:"bytes,2,opt,name=uid"`
	Groups   []string              `json:"groups,omitempty" protobuf:"bytes,3,rep,name=groups"`
	Extra    map[string]ExtraValue `json:"extra,omitempty" protobuf:"bytes,4,rep,name=extra"`
}

type ExtraValue []string
