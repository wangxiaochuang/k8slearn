package user

type Info interface {
	GetName() string
	GetUID() string
	GetGroups() []string

	GetExtra() map[string][]string
}

type DefaultInfo struct {
	Name   string
	UID    string
	Groups []string
	Extra  map[string][]string
}

func (i *DefaultInfo) GetName() string {
	return i.Name
}

func (i *DefaultInfo) GetUID() string {
	return i.UID
}

func (i *DefaultInfo) GetGroups() []string {
	return i.Groups
}

func (i *DefaultInfo) GetExtra() map[string][]string {
	return i.Extra
}

// well-known user and group names
const (
	SystemPrivilegedGroup = "system:masters"
	NodesGroup            = "system:nodes"
	MonitoringGroup       = "system:monitoring"
	AllUnauthenticated    = "system:unauthenticated"
	AllAuthenticated      = "system:authenticated"

	Anonymous     = "system:anonymous"
	APIServerUser = "system:apiserver"

	// core kubernetes process identities
	KubeProxy             = "system:kube-proxy"
	KubeControllerManager = "system:kube-controller-manager"
	KubeScheduler         = "system:kube-scheduler"
)
