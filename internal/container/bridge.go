package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	BRIDGE_METADATA_DIR = "/var/lib/gonett/bridges"
)

type BridgeMetadata struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	CreatedAt  string   `json:"created_at"`
	Interfaces []string `json:"interfaces"`
}

type BridgeManager struct{}

func NewBridgeManager() *BridgeManager {
	os.MkdirAll(BRIDGE_METADATA_DIR, 0755)
	return &BridgeManager{}
}

func saveBridgeMeta(meta BridgeMetadata) error {
	path := filepath.Join(BRIDGE_METADATA_DIR, meta.ID+".json")
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func loadBridgeMeta(id string) (*BridgeMetadata, error) {
	path := filepath.Join(BRIDGE_METADATA_DIR, id+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m BridgeMetadata
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func deleteBridgeMeta(id string) error {
	return os.Remove(filepath.Join(BRIDGE_METADATA_DIR, id+".json"))
}

func (bm *BridgeManager) Create(name, ns string) error {
	nsPath := filepath.Join(NETNS_BASE, ns)

	// Save old namespace
	origNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get current netns: %w", err)
	}
	defer origNS.Close()

	// Open target ns
	targetNS, err := netns.GetFromPath(nsPath)
	if err != nil {
		return fmt.Errorf("open target netns '%s': %w", ns, err)
	}
	defer targetNS.Close()

	// Set namespace
	if err := netns.Set(targetNS); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Create bridge inside target ns
	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
	}

	if err := netlink.LinkAdd(br); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("bridge add: %w", err)
	}

	// Bring up the bridge interface
	if err := netlink.LinkSetUp(br); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("bridge up: %w", err)
	}

	// Return to original namespace
	if err := netns.Set(origNS); err != nil {
		return fmt.Errorf("setns back: %w", err)
	}

	// Persist metadata
	id, _ := generateNamespaceID()
	meta := BridgeMetadata{
		ID:         id,
		Name:       name,
		Namespace:  ns,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Interfaces: []string{},
	}

	return saveBridgeMeta(meta)
}

func (bm *BridgeManager) AddInterface(bridgeID, ifName string) error {
	meta, err := loadBridgeMeta(bridgeID)
	if err != nil {
		return err
	}

	nsPath := filepath.Join(NETNS_BASE, meta.Namespace)

	// Save current ns
	origNS, _ := netns.Get()
	defer origNS.Close()

	// Enter target ns
	targetNS, err := netns.GetFromPath(nsPath)
	if err != nil {
		return fmt.Errorf("open bridge ns: %w", err)
	}
	defer targetNS.Close()

	if err := netns.Set(targetNS); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Lookup bridge
	brLink, err := netlink.LinkByName(meta.Name)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup bridge: %w", err)
	}

	// Lookup interface
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup iface %s: %w", ifName, err)
	}

	// Attach
	if err := netlink.LinkSetMaster(ifLink, brLink); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("add interface to bridge: %w", err)
	}

	// Bring port UP
	if err := netlink.LinkSetUp(ifLink); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("iface up: %w", err)
	}

	// Back to original namespace
	if err := netns.Set(origNS); err != nil {
		return err
	}

	// Update metadata
	meta.Interfaces = append(meta.Interfaces, ifName)
	return saveBridgeMeta(*meta)
}

func (bm *BridgeManager) Delete(id string) error {
	meta, err := loadBridgeMeta(id)
	if err != nil {
		return err
	}

	nsPath := filepath.Join(NETNS_BASE, meta.Namespace)

	// Save current ns
	origNS, _ := netns.Get()
	defer origNS.Close()

	// Enter target namespace
	targetNS, err := netns.GetFromPath(nsPath)
	if err != nil {
		return fmt.Errorf("get ns: %w", err)
	}
	defer targetNS.Close()

	if err := netns.Set(targetNS); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Delete bridge link
	if br, err := netlink.LinkByName(meta.Name); err == nil {
		netlink.LinkDel(br)
	}

	// Restore namespace
	netns.Set(origNS)

	return deleteBridgeMeta(id)
}

func (bm *BridgeManager) List() ([]BridgeMetadata, error) {
	files, err := os.ReadDir(BRIDGE_METADATA_DIR)
	if err != nil {
		return nil, err
	}

	var out []BridgeMetadata

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}

		id := f.Name()[:len(f.Name())-5]
		meta, err := loadBridgeMeta(id)
		if err == nil {
			out = append(out, *meta)
		}
	}
	return out, nil
}
