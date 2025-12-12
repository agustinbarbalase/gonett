package repository

import (
	"encoding/json"
	"fmt"
	"gonett/internal/container/domain"
	"gonett/internal/container/utils"
	"os"
	"path/filepath"
)

const VETH_METADATA_DIR = "/var/lib/gonett/veths"

type VethRepository struct {
	namespaceRepo *NamespaceRepository
}

func NewVethRepository(namespaceRepo *NamespaceRepository) *VethRepository {
	os.MkdirAll(VETH_METADATA_DIR, 0755)
	return &VethRepository{
		namespaceRepo: namespaceRepo,
	}
}

func (vr *VethRepository) Save(veth *domain.Veth) error {
	// Generate ID if not present
	if veth.ID == "" {
		id, err := utils.GenerateID()
		if err != nil {
			return fmt.Errorf("generate id: %w", err)
		}
		veth.ID = id
	}

	// Save namespaces if present
	if vr.namespaceRepo != nil {
		if veth.NamespaceA != nil {
			if err := vr.namespaceRepo.Save(veth.NamespaceA); err != nil {
				return fmt.Errorf("save namespace A: %w", err)
			}
		}
		if veth.NamespaceB != nil {
			if err := vr.namespaceRepo.Save(veth.NamespaceB); err != nil {
				return fmt.Errorf("save namespace B: %w", err)
			}
		}
	}

	// Save veth metadata
	path := filepath.Join(VETH_METADATA_DIR, veth.ID+".json")
	data, err := json.MarshalIndent(veth, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (vr *VethRepository) FindByID(id string) (*domain.Veth, error) {
	path := filepath.Join(VETH_METADATA_DIR, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var veth domain.Veth
	if err := json.Unmarshal(data, &veth); err != nil {
		return nil, err
	}

	return &veth, nil
}

func (vr *VethRepository) FindByName(name string) (*domain.Veth, error) {
	files, err := os.ReadDir(VETH_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(VETH_METADATA_DIR, file.Name()))
		if err != nil {
			continue
		}

		var veth domain.Veth
		if err := json.Unmarshal(data, &veth); err != nil {
			continue
		}

		if veth.Name == name {
			return &veth, nil
		}
	}

	return nil, fmt.Errorf("veth with name %s not found", name)
}

func (vr *VethRepository) List() ([]*domain.Veth, error) {
	files, err := os.ReadDir(VETH_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var veths []*domain.Veth
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		veth, err := vr.FindByID(id)
		if err == nil {
			veths = append(veths, veth)
		}
	}

	return veths, nil
}

func (vr *VethRepository) Delete(id string) error {
	// First, get the veth to know what to clean up
	if vr.namespaceRepo != nil {
		veth, err := vr.FindByID(id)
		if err == nil {
			// Delete namespaces if present
			if veth.NamespaceA != nil {
				vr.namespaceRepo.Delete(veth.NamespaceA.ID)
			}
			if veth.NamespaceB != nil {
				vr.namespaceRepo.Delete(veth.NamespaceB.ID)
			}
		}
	}

	// Delete veth metadata
	path := filepath.Join(VETH_METADATA_DIR, id+".json")
	return os.Remove(path)
}
