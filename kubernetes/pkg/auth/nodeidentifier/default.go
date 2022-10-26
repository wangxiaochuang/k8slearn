package nodeidentifier

import (
	"k8s.io/apiserver/pkg/authentication/user"
)

func NewDefaultNodeIdentifier() NodeIdentifier {
	return defaultNodeIdentifier{}
}

type defaultNodeIdentifier struct{}

const nodeUserNamePrefix = "system:node:"

func (defaultNodeIdentifier) NodeIdentity(u user.Info) (string, bool) {
	panic("not implemented")
}
