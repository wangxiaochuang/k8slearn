package admission

// p123
type Interface interface {
	Handles(operation Operation) bool
}

// p147
type Operation string

const (
	Create  Operation = "CREATE"
	Update  Operation = "UPDATE"
	Delete  Operation = "DELETE"
	Connect Operation = "CONNECT"
)

type PluginInitializer interface {
	Initialize(plugin Interface)
}
