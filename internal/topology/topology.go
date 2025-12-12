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
	})
}
