package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle internal nsenter command (used by attach)
	if command == "__gonett_nsenter__" {
		cmdNsenter()
		return
	}

	switch command {
	case "ls", "list":
		cmdList()
	case "rm", "remove":
		cmdRemove()
	case "attach":
		cmdAttach()
	case "exec":
		cmdExec()
	case "build":
		cmdBuild()
	case "cleanup":
		cmdCleanup()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("gonett - Container Network Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gonett ls                    List all containers")
	fmt.Println("  gonett rm <id>               Remove a container")
	fmt.Println("  gonett attach <id>           Attach to container shell")
	fmt.Println("  gonett exec <id> <command>   Execute command in container")
	fmt.Println("  gonett build                 Build topology from main.go")
	fmt.Println("  gonett cleanup               Remove all containers")
	fmt.Println("  gonett help                  Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gonett ls")
	fmt.Println("  gonett attach h1")
	fmt.Println("  gonett attach b819")
	fmt.Println("  gonett exec h1 ip addr show")
	fmt.Println("  gonett rm h1")
}
