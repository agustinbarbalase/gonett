package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gonett/internal/container/domain"
	"gonett/internal/container/manager"
	"gonett/internal/container/repository"
)

func cmdAttach() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gonett attach <container-id|name>")
		os.Exit(1)
	}

	target := os.Args[2]

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

	// Find container by ID or name
	var container *domain.Container
	for _, c := range containers {
		if strings.HasPrefix(c.ID, target) || c.Name == target {
			container = c
			break
		}
	}

	if container == nil {
		fmt.Printf("Container '%s' not found\n", target)
		os.Exit(1)
	}

	// Attach to container
	if err := cm.AttachContainer(container); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
