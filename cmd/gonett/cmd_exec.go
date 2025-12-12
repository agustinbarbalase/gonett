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

func cmdExec() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gonett exec <container-id|name> <command> [args...]")
		os.Exit(1)
	}

	target := os.Args[2]
	command := os.Args[3:]

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

	// Execute command
	if err := cm.ExecCommand(container, command); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
