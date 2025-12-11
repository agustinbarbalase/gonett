package nodes

import (
	nsManager "gonett/internal/namespace/manager"
)

type LinuxSwitch struct {
	name      string
	nsManager nsManager.LinuxNamespaceManager
}

func NewSwitch(name string) Switch {
	return LinuxSwitch{name: name}
}

func (h LinuxSwitch) GetName() string {
	return h.name
}

func (h LinuxSwitch) Link(switchNode Switch) error {
	// Implementation for linking host to switch goes here
	return nil
}

func (h LinuxSwitch) Delete() error {
	// Implementation for deleting the host goes here
	return nil
}
