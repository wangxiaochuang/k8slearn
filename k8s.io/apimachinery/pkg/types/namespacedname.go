package types

type NamespacedName struct {
	Namespace string
	Name      string
}

const (
	Separator = '/'
)

func (n NamespacedName) String() string {
	return n.Namespace + string(Separator) + n.Name
}
