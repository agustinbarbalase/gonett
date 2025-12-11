package namespace

type NamespaceManager interface {
	Create(name string) error
	Delete(name string) error
	List() ([]string, error)
}
