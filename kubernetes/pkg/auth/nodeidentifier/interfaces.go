package nodeidentifier

import (
	"k8s.io/apiserver/pkg/authentication/user"
)

type NodeIdentifier interface {
	NodeIdentity(user.Info) (nodeName string, isNode bool)
}
