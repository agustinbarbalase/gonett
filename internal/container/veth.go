package container

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	VETH_METADATA_DIR = "/var/lib/gonett/veths"
)

type VethMetadata struct {
	ID        string `json:"id"`
	NameA     string `json:"name_a"`
	NameB     string `json:"name_b"`
	CreatedAt string `json:"created_at"`
}

type VethManager struct{}

func NewVethManager() *VethManager {
	_ = os.MkdirAll(VETH_METADATA_DIR, 0755)
	return &VethManager{}
}

func generateVethID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func saveVethMetadata(meta VethMetadata) error {
	path := filepath.Join(VETH_METADATA_DIR, meta.ID+".json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func loadVethMetadata(id string) (*VethMetadata, error) {
	path := filepath.Join(VETH_METADATA_DIR, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta VethMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func deleteVethMetadata(id string) error {
	path := filepath.Join(VETH_METADATA_DIR, id+".json")
	return os.Remove(path)
}

func (vm *VethManager) Create(nameA, nameB string) error {
	v := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: nameA,
		},
		PeerName: nameB,
	}

	if err := netlink.LinkAdd(v); err != nil {
		return fmt.Errorf("create veth %s<->%s: %w", nameA, nameB, err)
	}

	id, err := generateVethID()
	if err != nil {
		return err
	}

	meta := &VethMetadata{
		ID:        id,
		NameA:     nameA,
		NameB:     nameB,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := saveVethMetadata(*meta); err != nil {
		return err
	}

	return nil
}

func (vm *VethManager) Delete(identifier string) error {
	entries, err := os.ReadDir(VETH_METADATA_DIR)
	if err != nil {
		return err
	}

	var meta *VethMetadata

	for _, file := range entries {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		m, err := loadVethMetadata(id)
		if err != nil {
			continue
		}

		if id == identifier {
			meta = m
			break
		}
	}

	if meta == nil {
		return fmt.Errorf("veth %s not found", identifier)
	}

	for _, ifname := range []string{meta.NameA, meta.NameB} {
		link, err := netlink.LinkByName(ifname)
		if err == nil {
			netlink.LinkDel(link)
		}
	}

	if err := deleteVethMetadata(meta.ID); err != nil {
		return err
	}

	return nil
}

func (vm *VethManager) List() ([]VethMetadata, error) {
	entries, err := os.ReadDir(VETH_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	result := []VethMetadata{}
	for _, file := range entries {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5]
		meta, err := loadVethMetadata(id)
		if err == nil {
			result = append(result, *meta)
		}
	}
	return result, nil
}

func (vm *VethManager) MoveEndToNamespace(ifname, ns string) error {
	link, err := netlink.LinkByName(ifname)
	if err != nil {
		return fmt.Errorf("move: %w", err)
	}

	nsPath := filepath.Join(NETNS_BASE, ns)
	f, err := os.Open(nsPath)
	if err != nil {
		return fmt.Errorf("open namespace %s: %w", ns, err)
	}
	defer f.Close()

	return netlink.LinkSetNsFd(link, int(f.Fd()))
}

func (vm *VethManager) SetUp(ifname string) error {
	link, err := netlink.LinkByName(ifname)
	if err != nil {
		return err
	}
	return netlink.LinkSetUp(link)
}
