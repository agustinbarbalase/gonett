package nodes

type Host interface {
	GetName() string
	Link(switchNode Switch) error
	Delete() error
}
