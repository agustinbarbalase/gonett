package domain

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vishvananda/netns"
)

const (
	NETNS_BASE = "/var/run/gonett/netns"
)

// Namespace represents a network namespace
type Namespace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Path      string `json:"path"`
	Runner    string `json:"runner"`
}

func CreateNamespace(name string) *Namespace {
	if err := os.MkdirAll(NETNS_BASE, 0755); err != nil {
		return nil
	}

	nsPath := filepath.Join(NETNS_BASE, name)

	// Create new named namespace
	ns, err := netns.NewNamed(name)
	if err != nil {
		return nil
	}
	ns.Close()

	namespace := &Namespace{
		ID:        "",
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
		Path:      filepath.Join("/var/run/netns", name),
		Runner:    nsPath,
	}

	return namespace
}

// Delete removes the network namespace
func (ns *Namespace) Delete() error {
	// Delete named namespace
	if err := netns.DeleteNamed(ns.Name); err != nil {
		return fmt.Errorf("delete netns: %w", err)
	}

	return nil
}
