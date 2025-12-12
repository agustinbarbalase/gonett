package topology

type NodeType string

const (
	NodeHost   NodeType = "host"
	NodeSwitch NodeType = "switch"
)

type Node struct {
	Name string
	Type NodeType
}

type Link struct {
	NodeA string
	NodeB string
	IPA   string // IP address for NodeA end (CIDR format, e.g., "10.0.0.1/24")
	IPB   string // IP address for NodeB end (CIDR format, e.g., "10.0.0.2/24")
}

type Topology struct {
	Nodes map[string]Node
	Links []Link
}

func NewTopology() *Topology {
	return &Topology{
		Nodes: map[string]Node{},
		Links: []Link{},
	}
}

func (t *Topology) AddHost(name string) {
	t.Nodes[name] = Node{
		Name: name,
		Type: NodeHost,
	}
}

func (t *Topology) AddSwitch(name string) {
	t.Nodes[name] = Node{
		Name: name,
		Type: NodeSwitch,
	}
}

func (t *Topology) AddLink(a, b string) {
	t.Links = append(t.Links, Link{
		NodeA: a,
		NodeB: b,
		IPA:   "",
		IPB:   "",
	})
}

func (t *Topology) AddLinkWithIPs(a, b, ipA, ipB string) {
	t.Links = append(t.Links, Link{
		NodeA: a,
		NodeB: b,
		IPA:   ipA,
		IPB:   ipB,
	})
}
