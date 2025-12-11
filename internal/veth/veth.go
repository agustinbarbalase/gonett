package veth

type vEthManager interface {
	Create(name string) error
	AssignToNamespace(vethName, namespace string) error
	Up(name string) error
	Delete(name string) error
}
