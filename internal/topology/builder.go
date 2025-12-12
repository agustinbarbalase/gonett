package topology

import (
	"fmt"

	"gonett/internal/container"
)

type Builder struct {
	nsMgr   *container.NamespaceManager
	vethMgr *container.VethManager
	brMgr   *container.BridgeManager
}

func NewBuilder() *Builder {
	return &Builder{
		nsMgr:   container.NewNamespaceManager(),
		vethMgr: container.NewVethManager(),
		brMgr:   container.NewBridgeManager(),
	}
}

func (b *Builder) Build(t *Topology) error {
	for _, node := range t.Nodes {
		if err := b.nsMgr.Create(node.Name); err != nil {
			return fmt.Errorf("create ns %s: %w", node.Name, err)
		}

		// Switches: create bridge *inside namespace*
		if node.Type == NodeSwitch {
			if err := b.brMgr.Create("br0", node.Name); err != nil {
				return fmt.Errorf("create bridge for switch %s: %w", node.Name, err)
			}
		}
	}

	for _, link := range t.Links {

		nodeA := t.Nodes[link.NodeA]
		nodeB := t.Nodes[link.NodeB]

		// Unique interface names
		ifA := fmt.Sprintf("%s-%s", nodeA.Name, nodeB.Name)
		ifB := fmt.Sprintf("%s-%s", nodeB.Name, nodeA.Name)

		// Create veth pair
		if err := b.vethMgr.Create(ifA, ifB); err != nil {
			return fmt.Errorf("veth %s<->%s: %w", ifA, ifB, err)
		}

		// If side B is a switch → attach to its bridge
		if nodeA.Type == NodeSwitch {
			if err := b.brMgr.AddInterface(b.findBridgeID(nodeA.Name), ifA); err != nil {
				return fmt.Errorf("attach %s to switch %s: %w", ifA, nodeA.Name, err)
			}
		}
		if nodeB.Type == NodeSwitch {
			if err := b.brMgr.AddInterface(b.findBridgeID(nodeB.Name), ifB); err != nil {
				return fmt.Errorf("attach %s to switch %s: %w", ifB, nodeB.Name, err)
			}
		}
	}

	return nil
}

func (b *Builder) findBridgeID(ns string) string {
	list, _ := b.brMgr.List()
	for _, br := range list {
		if br.Namespace == ns {
			return br.ID
		}
	}
	return ""
}
