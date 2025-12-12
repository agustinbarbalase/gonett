package repository

import (
	"encoding/json"
	"fmt"
	"gonett/internal/container/domain"
	"gonett/internal/container/utils"
	"os"
	"path/filepath"
)

const NAMESPACE_METADATA_DIR = "/var/lib/gonett/namespaces"

type NamespaceRepository struct{}

func NewNamespaceRepository() *NamespaceRepository {
	os.MkdirAll(NAMESPACE_METADATA_DIR, 0755)
	return &NamespaceRepository{}
}

func (nr *NamespaceRepository) Save(namespace *domain.Namespace) error {
	// Generate ID if not present
	if namespace.ID == "" {
		id, err := utils.GenerateID()
		if err != nil {
			return fmt.Errorf("generate id: %w", err)
		}
		namespace.ID = id
	}

	path := filepath.Join(NAMESPACE_METADATA_DIR, namespace.ID+".json")
	data, err := json.MarshalIndent(namespace, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (nr *NamespaceRepository) FindByID(id string) (*domain.Namespace, error) {
	path := filepath.Join(NAMESPACE_METADATA_DIR, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var namespace domain.Namespace
	if err := json.Unmarshal(data, &namespace); err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (nr *NamespaceRepository) FindByName(name string) (*domain.Namespace, error) {
	files, err := os.ReadDir(NAMESPACE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(NAMESPACE_METADATA_DIR, file.Name()))
		if err != nil {
			continue
		}

		var namespace domain.Namespace
		if err := json.Unmarshal(data, &namespace); err != nil {
			continue
		}

		if namespace.Name == name {
			return &namespace, nil
		}
	}

	return nil, fmt.Errorf("namespace with name %s not found", name)
}

func (nr *NamespaceRepository) List() ([]*domain.Namespace, error) {
	files, err := os.ReadDir(NAMESPACE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var namespaces []*domain.Namespace
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		namespace, err := nr.FindByID(id)
		if err == nil {
			namespaces = append(namespaces, namespace)
		}
	}

	return namespaces, nil
}

func (nr *NamespaceRepository) Delete(id string) error {
	path := filepath.Join(NAMESPACE_METADATA_DIR, id+".json")
	return os.Remove(path)
}
