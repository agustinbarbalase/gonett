package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"gonett/internal/namespace"
	nsmanager "gonett/internal/namespace/manager"
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
		// Verify that this was called legitimately from the parent process
		if os.Getenv("GONETT_CHILD_AUTH") != "true" {
			fmt.Println("unknown command")
			usage()
			os.Exit(0)
		}
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

	mgr := nsmanager.NewLinuxNamespaceManager()

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
		identifier := os.Args[2]

		// try to resolve identifier (id or name) to the namespace name using metadata
		targetName := identifier
		if metas, err := nsmanager.NewLinuxNamespaceManager().List(); err == nil {
			for _, m := range metas {
				if m.ID == identifier || m.Name == identifier {
					targetName = m.Name
					break
				}
			}
		}

		ns := namespace.NewLinuxNamespace(targetName, false)
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

		// print header and rows in tabular form similar to `docker ps`
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE ID\tNAME\tCREATED")

		for _, meta := range list {
			id := meta.ID
			name := meta.Name
			created := "unknown"
			if t, err := time.Parse(time.RFC3339, meta.CreatedAt); err == nil {
				created = humanizeDuration(time.Since(t)) + " ago"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\n", id, name, created)
		}
		w.Flush()

	default:
		fmt.Println("unknown command")
		usage()
	}
}

// humanizeDuration returns a short, human-friendly duration string.
func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		s := int(d.Seconds())
		if s <= 0 {
			return "just now"
		}
		return fmt.Sprintf("%ds", s)
	}
	if d < time.Hour {
		m := int(d.Minutes())
		return fmt.Sprintf("%dm", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		return fmt.Sprintf("%dh", h)
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo", months)
	}
	years := months / 12
	return fmt.Sprintf("%dy", years)
}
