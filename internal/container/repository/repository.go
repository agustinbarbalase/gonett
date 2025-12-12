package repository

// Repositories holds all repository instances
type Repositories struct {
	NamespaceRepo *NamespaceRepository
	BridgeRepo    *BridgeRepository
	VethRepo      *VethRepository
	ContainerRepo *ContainerRepository
}

// InitializeRepositories initializes all repositories with proper dependencies
func InitializeRepositories() (*Repositories, error) {
	// Create repositories in order of dependencies
	namespaceRepo := NewNamespaceRepository()
	vethRepo := NewVethRepository(namespaceRepo)
	bridgeRepo := NewBridgeRepository(vethRepo)
	containerRepo := NewContainerRepository(namespaceRepo, bridgeRepo, vethRepo)

	return &Repositories{
		NamespaceRepo: namespaceRepo,
		BridgeRepo:    bridgeRepo,
		VethRepo:      vethRepo,
		ContainerRepo: containerRepo,
	}, nil
}
