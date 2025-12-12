package main

import (
	"fmt"
	"log"

	"gonett/internal/container/manager"
	"gonett/internal/container/repository"
)

func cmdCleanup() {
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

	if len(containers) == 0 {
		fmt.Println("No containers to delete")
		return
	}

	fmt.Printf("Found %d containers to delete\n\n", len(containers))

	// Delete all containers
	for _, container := range containers {
		fmt.Printf("Deleting container '%s' (ID: %s)...\n", container.Name, container.ID[:12])

		if err := cm.DeleteContainer(container); err != nil {
			fmt.Printf("Error deleting container: %v\n", err)
			continue
		}

		fmt.Println("  ✓ Deleted")
	}

	fmt.Println("\n✓ Cleanup complete!")
}
