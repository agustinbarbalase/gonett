package namespace

type NamespaceMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Path      string `json:"path"`
}

type NamespaceManager interface {
	Create(name string) error
	Delete(id string) error
	List() ([]NamespaceMetadata, error)
}
