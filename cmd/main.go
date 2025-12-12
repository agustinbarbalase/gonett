package main

import (
	"gonett/internal/topology"
)

func main() {
	topo := topology.NewTopology()

	topo.AddHost("h1")
	topo.AddSwitch("s1")
	topo.AddHost("h2")

	topo.AddLink("h1", "s1")
	topo.AddLink("s1", "h2")

	builder := topology.NewBuilder()
	builder.Build(topo)
}
