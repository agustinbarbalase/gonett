package repository

import (
	"encoding/json"
	"fmt"
	"gonett/internal/container/domain"
	"gonett/internal/container/utils"
	"os"
	"path/filepath"
)

const BRIDGE_METADATA_DIR = "/var/lib/gonett/bridges"

type BridgeRepository struct {
	vethRepo *VethRepository
}

func NewBridgeRepository(vethRepo *VethRepository) *BridgeRepository {
	os.MkdirAll(BRIDGE_METADATA_DIR, 0755)
	return &BridgeRepository{
		vethRepo: vethRepo,
	}
}

func (br *BridgeRepository) Save(bridge *domain.Bridge) error {
	// Generate ID if not present
	if bridge.ID == "" {
		id, err := utils.GenerateID()
		if err != nil {
			return fmt.Errorf("generate id: %w", err)
		}
		bridge.ID = id
	}

	// Save all veths in the bridge
	if br.vethRepo != nil {
		for i := range bridge.Veths {
			if err := br.vethRepo.Save(&bridge.Veths[i]); err != nil {
				return fmt.Errorf("save veth %s: %w", bridge.Veths[i].ID, err)
			}
		}
	}

	// Save bridge metadata
	path := filepath.Join(BRIDGE_METADATA_DIR, bridge.ID+".json")
	data, err := json.MarshalIndent(bridge, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (br *BridgeRepository) FindByID(id string) (*domain.Bridge, error) {
	path := filepath.Join(BRIDGE_METADATA_DIR, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bridge domain.Bridge
	if err := json.Unmarshal(data, &bridge); err != nil {
		return nil, err
	}

	return &bridge, nil
}

func (br *BridgeRepository) FindByName(name string) (*domain.Bridge, error) {
	files, err := os.ReadDir(BRIDGE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(BRIDGE_METADATA_DIR, file.Name()))
		if err != nil {
			continue
		}

		var bridge domain.Bridge
		if err := json.Unmarshal(data, &bridge); err != nil {
			continue
		}

		if bridge.Name == name {
			return &bridge, nil
		}
	}

	return nil, fmt.Errorf("bridge with name %s not found", name)
}

func (br *BridgeRepository) FindByNamespace(namespace string) ([]*domain.Bridge, error) {
	files, err := os.ReadDir(BRIDGE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var bridges []*domain.Bridge
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(BRIDGE_METADATA_DIR, file.Name()))
		if err != nil {
			continue
		}

		var bridge domain.Bridge
		if err := json.Unmarshal(data, &bridge); err != nil {
			continue
		}

		if bridge.Namespace != nil && bridge.Namespace.Name == namespace {
			bridges = append(bridges, &bridge)
		}
	}

	return bridges, nil
}

func (br *BridgeRepository) List() ([]*domain.Bridge, error) {
	files, err := os.ReadDir(BRIDGE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var bridges []*domain.Bridge
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		bridge, err := br.FindByID(id)
		if err == nil {
			bridges = append(bridges, bridge)
		}
	}

	return bridges, nil
}

func (br *BridgeRepository) Delete(id string) error {
	// First, get the bridge to know what to clean up
	if br.vethRepo != nil {
		bridge, err := br.FindByID(id)
		if err == nil {
			// Delete all veths in the bridge
			for _, veth := range bridge.Veths {
				br.vethRepo.Delete(veth.ID)
			}
		}
	}

	// Delete bridge metadata
	path := filepath.Join(BRIDGE_METADATA_DIR, id+".json")
	return os.Remove(path)
}
