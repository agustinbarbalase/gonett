package repository

import (
	"encoding/json"
	"fmt"
	"gonett/internal/container/domain"
	"gonett/internal/container/utils"
	"os"
	"path/filepath"
)

const CONTAINER_METADATA_DIR = "/var/lib/gonett/containers"

type ContainerRepository struct {
	namespaceRepo *NamespaceRepository
	bridgeRepo    *BridgeRepository
	vethRepo      *VethRepository
}

func NewContainerRepository(namespaceRepo *NamespaceRepository, bridgeRepo *BridgeRepository, vethRepo *VethRepository) *ContainerRepository {
	os.MkdirAll(CONTAINER_METADATA_DIR, 0755)

	// Ensure repositories are initialized
	if namespaceRepo == nil {
		namespaceRepo = NewNamespaceRepository()
	}
	if vethRepo == nil {
		vethRepo = NewVethRepository(namespaceRepo)
	}
	if bridgeRepo == nil {
		bridgeRepo = NewBridgeRepository(vethRepo)
	}

	return &ContainerRepository{
		namespaceRepo: namespaceRepo,
		bridgeRepo:    bridgeRepo,
		vethRepo:      vethRepo,
	}
}

func (cr *ContainerRepository) Save(container *domain.Container) error {
	// Generate ID if not present
	if container.ID == "" {
		id, err := utils.GenerateID()
		if err != nil {
			return fmt.Errorf("generate id: %w", err)
		}
		container.ID = id
	}

	// Save namespace if present
	if container.Namespace != nil {
		if err := cr.namespaceRepo.Save(container.Namespace); err != nil {
			return fmt.Errorf("save namespace: %w", err)
		}
	}

	// Save all bridges
	for i := range container.Bridges {
		if err := cr.bridgeRepo.Save(&container.Bridges[i]); err != nil {
			return fmt.Errorf("save bridge %s: %w", container.Bridges[i].ID, err)
		}
	}

	// Save all veths
	for i := range container.Veths {
		if err := cr.vethRepo.Save(&container.Veths[i]); err != nil {
			return fmt.Errorf("save veth %s: %w", container.Veths[i].ID, err)
		}
	}

	// Save container metadata
	path := filepath.Join(CONTAINER_METADATA_DIR, container.ID+".json")
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (cr *ContainerRepository) FindByID(containerID string) (*domain.Container, error) {
	path := filepath.Join(CONTAINER_METADATA_DIR, containerID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var container domain.Container
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	return &container, nil
}

func (cr *ContainerRepository) FindByName(name string) (*domain.Container, error) {
	files, err := os.ReadDir(CONTAINER_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(CONTAINER_METADATA_DIR, file.Name()))
		if err != nil {
			continue
		}

		var container domain.Container
		if err := json.Unmarshal(data, &container); err != nil {
			continue
		}

		if container.Name == name {
			return &container, nil
		}
	}

	return nil, fmt.Errorf("container with name %s not found", name)
}

func (cr *ContainerRepository) List() ([]*domain.Container, error) {
	files, err := os.ReadDir(CONTAINER_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var containers []*domain.Container
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		container, err := cr.FindByID(id)
		if err == nil {
			containers = append(containers, container)
		}
	}

	return containers, nil
}

func (cr *ContainerRepository) Delete(containerID string) error {
	// First, get the container to know what to clean up
	container, err := cr.FindByID(containerID)
	if err != nil {
		return err
	}

	// Delete namespace if present
	if container.Namespace != nil {
		if err := cr.namespaceRepo.Delete(container.Namespace.ID); err != nil {
			// Log but don't fail if namespace delete fails
		}
	}

	// Delete all bridges
	for _, bridge := range container.Bridges {
		if err := cr.bridgeRepo.Delete(bridge.ID); err != nil {
			// Log but don't fail
		}
	}

	// Delete all veths
	for _, veth := range container.Veths {
		if err := cr.vethRepo.Delete(veth.ID); err != nil {
			// Log but don't fail
		}
	}

	// Delete container metadata
	path := filepath.Join(CONTAINER_METADATA_DIR, containerID+".json")
	return os.Remove(path)
}
