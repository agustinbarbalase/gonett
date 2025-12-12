package container

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

const (
	NETNS_BASE   = "/var/run/gonett/netns"
	NAMESPACE_METADATA_DIR = "/var/lib/gonett/namespaces"
)

type NamespaceMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Path      string `json:"path"`
}

func saveMetadata(meta NamespaceMetadata) error {
	path := filepath.Join(NAMESPACE_METADATA_DIR, meta.ID+".json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func loadMetadata(id string) (*NamespaceMetadata, error) {
	path := filepath.Join(NAMESPACE_METADATA_DIR, id+".json")
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
	path := filepath.Join(NAMESPACE_METADATA_DIR, id+".json")
	return os.Remove(path)
}

type NamespaceManager struct{}

func createNamespaceDirectory() error {
	_, err := os.Stat(NETNS_BASE)
	if os.IsNotExist(err) {
		return os.MkdirAll(NETNS_BASE, 0755)
	}
	return err
}

func createMetaDataDirectory() error {
	_, err := os.Stat(NAMESPACE_METADATA_DIR)
	if os.IsNotExist(err) {
		return os.MkdirAll(NAMESPACE_METADATA_DIR, 0755)
	}
	return err
}

func generateNamespaceID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewNamespaceManager() *NamespaceManager {
	createNamespaceDirectory()
	createMetaDataDirectory()
	return &NamespaceManager{}
}

func (ln *NamespaceManager) Create(name string) error {
	if err := os.MkdirAll(NETNS_BASE, 0755); err != nil {
		return err
	}

	nsPath := filepath.Join(NETNS_BASE, name)

	f, err := os.Create(nsPath)
	if err != nil {
		return err
	}
	f.Close()

	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("unshare: %w", err)
	}

	if err := unix.Mount("/proc/self/ns/net", nsPath, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("bind mount: %w", err)
	}

	id, err := generateNamespaceID()
	if err != nil {
		return err
	}

	meta := NamespaceMetadata{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
		Path:      nsPath,
	}

	if err := saveMetadata(meta); err != nil {
		return err
	}

	return nil
}

func (ln *NamespaceManager) Delete(identifier string) error {
	entries, err := os.ReadDir(NAMESPACE_METADATA_DIR)
	if err != nil {
		return err
	}

	var meta *NamespaceMetadata

	for _, file := range entries {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		m, err := loadMetadata(id)
		if err != nil {
			continue
		}

		if id == identifier {
			meta = m
			break
		}
	}

	if meta == nil {
		return fmt.Errorf("no container found with id or name: %s", identifier)
	}

	if err := unix.Unmount(meta.Path, 0); err != nil {
		return fmt.Errorf("umount: %w", err)
	}

	if err := os.Remove(meta.Path); err != nil {
		return fmt.Errorf("rm ns: %w", err)
	}

	if err := deleteMetadata(meta.ID); err != nil {
		return fmt.Errorf("rm metadata: %w", err)
	}

	return nil
}

func (ln *NamespaceManager) List() ([]NamespaceMetadata, error) {
	entries, err := os.ReadDir(NAMESPACE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	containers := []NamespaceMetadata{}
	for _, file := range entries {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5] // remove `.json`
		meta, err := loadMetadata(id)
		if err == nil {
			containers = append(containers, *meta)
		}
	}

	return containers, nil
}
