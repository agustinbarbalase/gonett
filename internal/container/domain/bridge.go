package domain

import (
	"fmt"
	"runtime"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// Bridge represents a network bridge
type Bridge struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Namespace *Namespace `json:"namespace,omitempty"`
	CreatedAt string     `json:"created_at"`
	Veths     []Veth     `json:"veths,omitempty"`
}

func NewBridge(id, name string, namespace *Namespace) *Bridge {
	return &Bridge{
		ID:        id,
		Name:      name,
		Namespace: namespace,
		CreatedAt: time.Now().Format(time.RFC3339),
		Veths:     []Veth{},
	}
}

func CreateBridge(name string, namespace *Namespace) (*Bridge, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("get current netns: %w", err)
	}
	defer origNS.Close()

	// Open target ns
	targetNS, err := netns.GetFromPath(namespace.Path)
	if err != nil {
		return nil, fmt.Errorf("open target netns '%s': %w", namespace.Name, err)
	}
	defer targetNS.Close()

	// Set namespace
	if err := netns.Set(targetNS); err != nil {
		return nil, fmt.Errorf("setns: %w", err)
	}

	// Create bridge inside target ns
	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
	}

	if err := netlink.LinkAdd(br); err != nil {
		netns.Set(origNS)
		return nil, fmt.Errorf("bridge add: %w", err)
	}

	// Look up the bridge we just created to get a fresh handle
	link, err := netlink.LinkByName(name)
	if err != nil {
		netns.Set(origNS)
		return nil, fmt.Errorf("lookup bridge: %w", err)
	}

	// Bring up the bridge interface
	if err := netlink.LinkSetUp(link); err != nil {
		netns.Set(origNS)
		return nil, fmt.Errorf("bridge up: %w", err)
	}

	// Restore namespace immediately
	if err := netns.Set(origNS); err != nil {
		return nil, fmt.Errorf("setns back: %w", err)
	}

	bridge := &Bridge{
		ID:        "",
		Name:      name,
		Namespace: namespace,
		CreatedAt: time.Now().Format(time.RFC3339),
		Veths:     []Veth{},
	}

	return bridge, nil
}

// AddInterface adds an interface to the bridge
func (b *Bridge) AddInterface(veth Veth) error {
	// Save current ns
	origNS, _ := netns.Get()
	defer origNS.Close()

	// Enter target ns
	targetNS, err := netns.GetFromPath(b.Namespace.Path)
	if err != nil {
		return fmt.Errorf("open bridge ns: %w", err)
	}
	defer targetNS.Close()

	if err := netns.Set(targetNS); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Lookup bridge
	brLink, err := netlink.LinkByName(b.Name)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup bridge: %w", err)
	}

	// Lookup interface
	ifLink, err := netlink.LinkByName(veth.NamespaceA.Name)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup iface %s: %w", veth.NamespaceA.Name, err)
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

	b.Veths = append(b.Veths, veth)

	return nil
}

// AttachInterfaceByName attaches an interface to the bridge by interface name
func (b *Bridge) AttachInterfaceByName(ifName string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current ns
	origNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get current ns: %w", err)
	}
	defer origNS.Close()

	// Enter target ns
	targetNS, err := netns.GetFromPath(b.Namespace.Path)
	if err != nil {
		return fmt.Errorf("open namespace: %w", err)
	}
	defer targetNS.Close()

	if err := netns.Set(targetNS); err != nil {
		return fmt.Errorf("set namespace: %w", err)
	}

	// Lookup bridge
	brLink, err := netlink.LinkByName(b.Name)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup bridge %s: %w", b.Name, err)
	}

	// Lookup interface - get fresh handle
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup interface %s: %w", ifName, err)
	}

	// Attach to bridge
	if err := netlink.LinkSetMaster(ifLink, brLink); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("set master: %w", err)
	}

	// Get fresh link handle for bringing up
	ifLink, err = netlink.LinkByName(ifName)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("lookup interface for up: %w", err)
	}

	// Bring interface up
	if err := netlink.LinkSetUp(ifLink); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("set up: %w", err)
	}

	// Restore namespace immediately
	netns.Set(origNS)

	return nil
}

// Delete removes the bridge
func (b *Bridge) Delete() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current ns
	origNS, _ := netns.Get()
	defer origNS.Close()

	// Enter target namespace (skip if namespace no longer exists)
	targetNS, err := netns.GetFromPath(b.Namespace.Path)
	if err != nil {
		// Namespace already deleted, nothing to clean up
		return nil
	}
	defer targetNS.Close()

	if err := netns.Set(targetNS); err != nil {
		netns.Set(origNS)
		return nil // Namespace gone, skip cleanup
	}

	// Delete bridge link
	if br, err := netlink.LinkByName(b.Name); err == nil {
		netlink.LinkDel(br)
	}

	// Restore namespace
	netns.Set(origNS)

	return nil
}
