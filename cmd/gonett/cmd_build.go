package main

import (
	"log"

	"gonett/internal/topology"
)

func cmdBuild() {
	topo := topology.NewTopology()
	topo.AddHost("h1")
	topo.AddHost("h2")
	topo.AddSwitch("s1")
	topo.AddLinkWithIPs("h1", "s1", "10.0.0.1/24", "")
	topo.AddLinkWithIPs("h2", "s1", "10.0.0.2/24", "")

	// Build it
	builder, err := topology.NewBuilder()
	if err != nil {
		log.Fatalf("Failed to create builder: %v", err)
	}

	if err := builder.Build(topo); err != nil {
		log.Fatalf("Failed to build topology: %v", err)
	}
}
