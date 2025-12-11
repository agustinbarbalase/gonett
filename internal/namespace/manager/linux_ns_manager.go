package namespace

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

const (
	NETNS_BASE   = "/var/run/gonett/netns"
	METADATA_DIR = "/var/lib/gonett/namespaces"
)

type LinuxNamespaceManager struct{}

func createNamespaceDirectory() error {
	_, err := os.Stat(NETNS_BASE)
	if os.IsNotExist(err) {
		return os.MkdirAll(NETNS_BASE, 0755)
	}
	return err
}

func createMetaDataDirectory() error {
	_, err := os.Stat(METADATA_DIR)
	if os.IsNotExist(err) {
		return os.MkdirAll(METADATA_DIR, 0755)
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

func NewLinuxNamespaceManager() NamespaceManager {
	createNamespaceDirectory()
	createMetaDataDirectory()
	return &LinuxNamespaceManager{}
}

func (ln *LinuxNamespaceManager) Create(name string) error {
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

func (ln *LinuxNamespaceManager) Delete(identifier string) error {
	entries, err := os.ReadDir(METADATA_DIR)
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

func (ln *LinuxNamespaceManager) List() ([]NamespaceMetadata, error) {
	entries, err := os.ReadDir(METADATA_DIR)
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
