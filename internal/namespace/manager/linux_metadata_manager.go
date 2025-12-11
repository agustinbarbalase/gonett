package namespace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func saveMetadata(meta NamespaceMetadata) error {
	path := filepath.Join(METADATA_DIR, meta.ID+".json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func loadMetadata(id string) (*NamespaceMetadata, error) {
	path := filepath.Join(METADATA_DIR, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta NamespaceMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func deleteMetadata(id string) error {
	path := filepath.Join(METADATA_DIR, id+".json")
	return os.Remove(path)
}
