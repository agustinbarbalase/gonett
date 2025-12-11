package nodes

import (
	nsManager "gonett/internal/namespace/manager"
)

type LinuxHost struct {
	name      string
	nsManager nsManager.LinuxNamespaceManager
}

func NewHost(name string) Host {
	return LinuxHost{name: name}
}

func (h LinuxHost) GetName() string {
	return h.name
}

func (h LinuxHost) Link(switchNode Switch) error {
	// Implementation for linking host to switch goes here
	return nil
}

func (h LinuxHost) Delete() error {
	// Implementation for deleting the host goes here
	return nil
}
