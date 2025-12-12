package manager

import (
	"fmt"
	"strings"

	"gonett/internal/container/domain"
	"gonett/internal/container/repository"
)

type ContainerManager struct {
	containerRepo *repository.ContainerRepository
	namespaceRepo *repository.NamespaceRepository
	bridgeRepo    *repository.BridgeRepository
	vethRepo      *repository.VethRepository
}

func NewContainerManager(
	containerRepo *repository.ContainerRepository,
	namespaceRepo *repository.NamespaceRepository,
	bridgeRepo *repository.BridgeRepository,
	vethRepo *repository.VethRepository,
) *ContainerManager {
	return &ContainerManager{
		containerRepo: containerRepo,
		namespaceRepo: namespaceRepo,
		bridgeRepo:    bridgeRepo,
		vethRepo:      vethRepo,
	}
}

// CreateContainer creates a new container with namespace
func (cm *ContainerManager) CreateContainer(name string) (*domain.Container, error) {
	container, err := domain.CreateWithNamespace(name)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	if err := cm.containerRepo.Save(container); err != nil {
		return nil, fmt.Errorf("save container: %w", err)
	}

	return container, nil
}

// ListContainers lists all containers
func (cm *ContainerManager) ListContainers() ([]*domain.Container, error) {
	containers, err := cm.containerRepo.List()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	if len(containers) == 0 {
		return containers, nil
	}

	return containers, nil
}

// AttachContainer attaches to a container shell
func (cm *ContainerManager) AttachContainer(container *domain.Container) error {
	if container.Namespace == nil {
		return fmt.Errorf("container has no namespace")
	}

	if err := container.AttachShell(); err != nil {
		return fmt.Errorf("attach to container: %w", err)
	}

	return nil
}

// ExecCommand executes a command in container
func (cm *ContainerManager) ExecCommand(container *domain.Container, cmd []string) error {
	if container.Namespace == nil {
		return fmt.Errorf("container has no namespace")
	}

	fmt.Printf("Executing: %s\n", strings.Join(cmd, " "))

	if err := container.Exec(cmd); err != nil {
		return fmt.Errorf("exec command: %w", err)
	}

	return nil
}

// DeleteContainer removes a container and its resources
func (cm *ContainerManager) DeleteContainer(container *domain.Container) error {
	fmt.Printf("\nDeleting container '%s'...\n", container.Name)

	// Delete namespace (which will cascade to cleanup)
	if container.Namespace != nil {
		if err := container.Namespace.Delete(); err != nil {
			fmt.Printf("Warning: failed to delete namespace: %v\n", err)
		}
	}

	// Delete bridges
	for _, bridge := range container.Bridges {
		if err := bridge.Delete(); err != nil {
			fmt.Printf("Warning: failed to delete bridge %s: %v\n", bridge.Name, err)
		}
	}

	// Delete veths
	for _, veth := range container.Veths {
		if err := veth.Delete(); err != nil {
			fmt.Printf("Warning: failed to delete veth %s: %v\n", veth.Name, err)
		}
	}

	// Delete from repository
	if err := cm.containerRepo.Delete(container.ID); err != nil {
		return fmt.Errorf("delete container: %w", err)
	}

	return nil
}

// CreateBridgeToContainer adds a bridge to an existing container
func (cm *ContainerManager) CreateBridgeToContainer(container *domain.Container, name string) (*domain.Bridge, error) {
	if container.Namespace == nil {
		return nil, fmt.Errorf("container has no namespace")
	}

	bridge, err := container.AddBridge(name)
	if err != nil {
		return nil, fmt.Errorf("add bridge: %w", err)
	}

	// Save the updated container
	if err := cm.containerRepo.Save(container); err != nil {
		return nil, fmt.Errorf("save container: %w", err)
	}

	return bridge, nil
}

// CreateInterfaceToContainer adds a veth pair to an existing container
func (cm *ContainerManager) CreateInterfaceToContainer(container *domain.Container, nameA, nameB string) (*domain.Veth, error) {
	if container.Namespace == nil {
		return nil, fmt.Errorf("container has no namespace")
	}

	veth, err := container.AddVeth(container.Namespace, container.Namespace, nameA)
	if err != nil {
		return nil, fmt.Errorf("add veth: %w", err)
	}

	// Save the updated container
	if err := cm.containerRepo.Save(container); err != nil {
		return nil, fmt.Errorf("save container: %w", err)
	}

	return veth, nil
}

// AddInterfaceToBridge connects a veth interface to a bridge in a container
func (cm *ContainerManager) AddInterfaceToBridge(container *domain.Container, vethName, bridgeName, vethEnd string) error {
	if container.Namespace == nil {
		return fmt.Errorf("container has no namespace")
	}

	if err := container.ConnectVethToBridge(vethName, bridgeName, vethEnd); err != nil {
		return fmt.Errorf("connect veth to bridge: %w", err)
	}

	// Save the updated container
	if err := cm.containerRepo.Save(container); err != nil {
		return fmt.Errorf("save container: %w", err)
	}

	return nil
}
