package domain

import (
	"fmt"
	"runtime"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// Veth represents a virtual ethernet pair
type Veth struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	PeerName   string     `json:"peer_name"`
	NamespaceA *Namespace `json:"namespace_a,omitempty"`
	NamespaceB *Namespace `json:"namespace_b,omitempty"`
	CreatedAt  string     `json:"created_at"`
}

// CreateVeth creates a new virtual ethernet pair
func CreateVeth(NamespaceA *Namespace, NamespaceB *Namespace, nameA string) (*Veth, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if nameA == "" {
		nameA = NamespaceA.Name + "-eth0"
	}
	// Generate peer name based on nameA to ensure uniqueness
	nameB := nameA + "-peer"
	if nameA == NamespaceA.Name+"-eth0" {
		nameB = NamespaceB.Name + "-eth0"
	}

	// Save current namespace and switch to init namespace for veth creation
	origNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("get current ns: %w", err)
	}
	defer origNS.Close()

	// Switch to init namespace to create veth pair
	initNS, err := netns.GetFromPath("/proc/1/ns/net")
	if err != nil {
		return nil, fmt.Errorf("get init ns: %w", err)
	}
	defer initNS.Close()

	if err := netns.Set(initNS); err != nil {
		return nil, fmt.Errorf("set to init ns: %w", err)
	}

	v := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: nameA,
		},
		PeerName: nameB,
	}

	if err := netlink.LinkAdd(v); err != nil {
		netns.Set(origNS)
		return nil, fmt.Errorf("create veth: create veth %s<->%s: %w", nameA, nameB, err)
	}

	// Restore namespace immediately
	netns.Set(origNS)

	veth := &Veth{
		ID:         "",
		Name:       nameA,
		PeerName:   nameB,
		NamespaceA: NamespaceA,
		NamespaceB: NamespaceB,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	return veth, nil
}

// MoveEndToNamespace moves one end of the veth to a namespace
func (v *Veth) MoveEndToNamespace(ifname string, namespace *Namespace) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current namespace
	origNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get current ns: %w", err)
	}
	defer origNS.Close()

	// Ensure we're in the default/init namespace to look up the interface
	initNS, err := netns.GetFromPath("/proc/1/ns/net")
	if err != nil {
		return fmt.Errorf("get init ns: %w", err)
	}
	defer initNS.Close()

	if err := netns.Set(initNS); err != nil {
		return fmt.Errorf("set to init ns: %w", err)
	}

	// Look up interface in the init namespace
	link, err := netlink.LinkByName(ifname)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("find interface %s: %w", ifname, err)
	}

	// Open target namespace
	nsHandle, err := netns.GetFromPath(namespace.Path)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("open namespace %s: %w", namespace.Name, err)
	}
	defer nsHandle.Close()

	// Move interface to namespace (while still in init namespace)
	if err := netlink.LinkSetNsFd(link, int(nsHandle)); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("set netns for %s: %w", ifname, err)
	}

	// Restore namespace after move is complete
	netns.Set(origNS)
	return nil
}

// AssignIP assigns an IP address to a veth interface inside a namespace
func (v *Veth) AssignIP(ifname, ipCIDR string, namespace *Namespace) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current namespace
	origNS, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get current ns: %w", err)
	}
	defer origNS.Close()

	// Open target namespace
	targetNS, err := netns.GetFromPath(namespace.Path)
	if err != nil {
		return fmt.Errorf("open namespace: %w", err)
	}
	defer targetNS.Close()

	// Set to target namespace
	if err := netns.Set(targetNS); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("set namespace: %w", err)
	}

	// Get link
	link, err := netlink.LinkByName(ifname)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("get link: %w", err)
	}

	// Parse IP address
	addr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		netns.Set(origNS)
		return fmt.Errorf("parse addr: %w", err)
	}

	// Add address to interface
	if err := netlink.AddrAdd(link, addr); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("add addr: %w", err)
	}

	// Bring interface up
	if err := netlink.LinkSetUp(link); err != nil {
		netns.Set(origNS)
		return fmt.Errorf("set up: %w", err)
	}

	// Restore namespace immediately
	netns.Set(origNS)
	return nil
}

func (v *Veth) Delete() error {
	link, err := netlink.LinkByName(v.Name)
	if err == nil {
		netlink.LinkDel(link)
	}

	return nil
}
