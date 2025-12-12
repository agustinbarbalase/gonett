package main

import (
	"fmt"
	"log"

	"gonett/internal/container/manager"
	"gonett/internal/container/repository"
)

func cmdList() {
	// Initialize repositories
	repos, err := repository.InitializeRepositories()
	if err != nil {
		log.Fatalf("Failed to initialize repositories: %v", err)
	}

	// Create container manager
	cm := manager.NewContainerManager(
		repos.ContainerRepo,
		repos.NamespaceRepo,
		repos.BridgeRepo,
		repos.VethRepo,
	)

	// Get all containers
	containers, err := cm.ListContainers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display containers in Docker-like format
	fmt.Printf("%-12s  %-20s  %-20s  %-8s  %-8s  %s\n",
		"CONTAINER ID", "NAME", "NAMESPACE", "BRIDGES", "VETHS", "CREATED")

	for _, c := range containers {
		containerID := c.ID
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}

		namespaceName := "-"
		if c.Namespace != nil {
			namespaceName = c.Namespace.Name
		}

		fmt.Printf("%-12s  %-20s  %-20s  %-8d  %-8d  %s\n",
			containerID,
			c.Name,
			namespaceName,
			len(c.Bridges),
			len(c.Veths),
			c.CreatedAt,
		)
	}
}
