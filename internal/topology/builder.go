package topology

import (
	"fmt"

	"gonett/internal/container/domain"
	"gonett/internal/container/manager"
	"gonett/internal/container/repository"
)

type Builder struct {
	cm            *manager.ContainerManager
	containerRepo *repository.ContainerRepository
	topology      *Topology
}

func NewBuilder() (*Builder, error) {
	// Initialize repositories
	repos, err := repository.InitializeRepositories()
	if err != nil {
		return nil, fmt.Errorf("initialize repositories: %w", err)
	}

	// Create container manager
	cm := manager.NewContainerManager(
		repos.ContainerRepo,
		repos.NamespaceRepo,
		repos.BridgeRepo,
		repos.VethRepo,
	)

	return &Builder{
		cm:            cm,
		containerRepo: repos.ContainerRepo,
		topology:      nil,
	}, nil
}

// Build creates containers for all nodes in the topology
func (b *Builder) Build(t *Topology) error {
	b.topology = t // Store topology for later reference

	// Create containers for each node
	nodeContainers := make(map[string]*domain.Container)

	for nodeName, node := range t.Nodes {
		var container *domain.Container
		var err error

		switch node.Type {
		case NodeHost:
			container, err = b.buildHost(nodeName)
		case NodeSwitch:
			container, err = b.buildSwitch(nodeName)
		default:
			return fmt.Errorf("unknown node type: %s", node.Type)
		}

		if err != nil {
			return fmt.Errorf("build node %s: %w", nodeName, err)
		}

		nodeContainers[nodeName] = container
	}

	// Create links between nodes
	for _, link := range t.Links {
		if err := b.buildLink(nodeContainers, link); err != nil {
			return fmt.Errorf("build link %s-%s: %w", link.NodeA, link.NodeB, err)
		}
	}

	fmt.Println("\n✓ Topology built successfully!")
	return nil
}

// buildHost creates a container for a host node
func (b *Builder) buildHost(name string) (*domain.Container, error) {
	fmt.Printf("\n  Creating host '%s'...\n", name)

	container, err := b.cm.CreateContainer(name)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	fmt.Printf("  ✓ Host '%s' created\n", name)
	return container, nil
}

// buildSwitch creates a container for a switch node with a bridge
func (b *Builder) buildSwitch(name string) (*domain.Container, error) {
	fmt.Printf("\n  Creating switch '%s'...\n", name)

	// Create container
	container, err := b.cm.CreateContainer(name)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	// Create bridge inside the switch container
	bridge, err := b.cm.CreateBridgeToContainer(container, fmt.Sprintf("%s-br0", name))
	if err != nil {
		return nil, fmt.Errorf("create bridge: %w", err)
	}

	fmt.Printf("  ✓ Switch '%s' created with bridge '%s'\n", name, bridge.Name)
	return container, nil
}

// buildLink creates a veth pair connecting two nodes
func (b *Builder) buildLink(nodeContainers map[string]*domain.Container, link Link) error {
	containerA := nodeContainers[link.NodeA]
	containerB := nodeContainers[link.NodeB]

	if containerA == nil || containerB == nil {
		return fmt.Errorf("missing container for link %s-%s", link.NodeA, link.NodeB)
	}

	// Generate interface names
	ifNameA := fmt.Sprintf("veth-%s-%s", link.NodeA, link.NodeB)

	fmt.Printf("  Creating link %s <--> %s\n", link.NodeA, link.NodeB)

	// Create veth pair connecting both containers
	veth, err := containerA.AddVeth(containerA.Namespace, containerB.Namespace, ifNameA)
	if err != nil {
		return fmt.Errorf("create veth pair: %w", err)
	}

	// Move each veth end to its target namespace
	if err := veth.MoveEndToNamespace(veth.Name, containerA.Namespace); err != nil {
		return fmt.Errorf("move veth name end: %w", err)
	}
	if err := veth.MoveEndToNamespace(veth.PeerName, containerB.Namespace); err != nil {
		return fmt.Errorf("move veth peer end: %w", err)
	}

	// Also add to containerB's veths
	containerB.Veths = append(containerB.Veths, *veth)

	// Save both containers
	if err := b.containerRepo.Save(containerA); err != nil {
		return fmt.Errorf("save container A: %w", err)
	}
	if err := b.containerRepo.Save(containerB); err != nil {
		return fmt.Errorf("save container B: %w", err)
	}

	// Now attach to bridges if needed (veths are already in namespaces)
	// If node B is a switch, attach the peer end to its bridge
	nodeB := b.getNodeByName(link.NodeB)
	if nodeB != nil && nodeB.Type == NodeSwitch {
		// Find the bridge in containerB
		for i := range containerB.Bridges {
			bridge := &containerB.Bridges[i]
			if err := bridge.AttachInterfaceByName(veth.PeerName); err != nil {
				return fmt.Errorf("attach peer to bridge: %w", err)
			}
		}
	}

	// If node A is a switch, attach the name end to its bridge
	nodeA := b.getNodeByName(link.NodeA)
	if nodeA != nil && nodeA.Type == NodeSwitch {
		// Find the bridge in containerA
		for i := range containerA.Bridges {
			bridge := &containerA.Bridges[i]
			if err := bridge.AttachInterfaceByName(veth.Name); err != nil {
				return fmt.Errorf("attach name to bridge: %w", err)
			}
		}
	}

	// Assign IP addresses if provided and node is a host
	if link.IPA != "" && nodeA != nil && nodeA.Type == NodeHost {
		if err := veth.AssignIP(veth.Name, link.IPA, containerA.Namespace); err != nil {
			return fmt.Errorf("assign IP to %s: %w", link.NodeA, err)
		}
		fmt.Printf("    IP %s assigned to %s\n", link.IPA, link.NodeA)
	}

	if link.IPB != "" && nodeB != nil && nodeB.Type == NodeHost {
		if err := veth.AssignIP(veth.PeerName, link.IPB, containerB.Namespace); err != nil {
			return fmt.Errorf("assign IP to %s: %w", link.NodeB, err)
		}
		fmt.Printf("    IP %s assigned to %s\n", link.IPB, link.NodeB)
	}

	fmt.Printf("  ✓ Link created: %s <--> %s\n", link.NodeA, link.NodeB)
	return nil
}

// getNodeByName helper to retrieve a node from topology
func (b *Builder) getNodeByName(name string) *Node {
	// This is a temporary implementation - in a real scenario, topology would be stored in Builder
	if b.topology == nil {
		return nil
	}
	if node, exists := b.topology.Nodes[name]; exists {
		return &node
	}
	return nil
}
