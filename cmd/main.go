package main

import (
	"fmt"
	"os"

	"gonet/internal/namespace"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd create <name>")
	fmt.Println("  go run ./cmd exec <name>")
	fmt.Println("  go run ./cmd delete <name>")
	fmt.Println("  go run ./cmd list")
}

func main() {
	// handle child mode (started via /proc/self/exe child <name>)
	if len(os.Args) > 1 && os.Args[1] == "child" {
		if len(os.Args) < 3 {
			fmt.Println("child: missing namespace name")
			os.Exit(1)
		}
		ns := namespace.NewLinuxNamespace(os.Args[2], true)
		if err := ns.Attach(); err != nil {
			fmt.Println("child attach error:", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd := os.Args[1]

	mgr := namespace.NewLinuxNamespaceManager()

	switch cmd {
	case "create":
		if len(os.Args) < 3 {
			usage()
			return
		}
		name := os.Args[2]
		if err := mgr.Create(name); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("namespace created:", name)

	case "exec":
		if len(os.Args) < 3 {
			usage()
			return
		}
		name := os.Args[2]
		// try to call Execute if the concrete manager exposes it
		if execer, ok := mgr.(interface{ Execute(string, []string) error }); ok {
			if err := execer.Execute(name, nil); err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			return
		}
		// fallback: attach a Namespace directly
		ns := namespace.NewLinuxNamespace(name, false)
		if err := ns.Attach(); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

	case "delete":
		if len(os.Args) < 3 {
			usage()
			return
		}
		name := os.Args[2]
		if err := mgr.Delete(name); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("namespace deleted:", name)

	case "list":
		list, err := mgr.List()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		for _, ns := range list {
			fmt.Println(ns)
		}

	default:
		fmt.Println("unknown command")
		usage()
	}
}
