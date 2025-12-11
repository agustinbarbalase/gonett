package nodes

type Switch interface {
	GetName() string
	Link(switchNode Switch) error
	Delete() error
}
