package domain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// Container represents a container with its network components
type Container struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt string     `json:"created_at"`
	Namespace *Namespace `json:"namespace,omitempty"`
	Bridges   []Bridge   `json:"bridges,omitempty"`
	Veths     []Veth     `json:"veths,omitempty"`
	isChild   bool       `json:"-"`
}

func NewContainer(name string, isChild bool) *Container {
	return &Container{
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
		isChild:   isChild,
		Bridges:   []Bridge{},
		Veths:     []Veth{},
	}
}

// CreateWithNamespace creates a new container and initializes its namespace
func CreateWithNamespace(name string) (*Container, error) {
	container := NewContainer(name, false)

	// Create namespace for the container
	ns := CreateNamespace(name)
	if ns == nil {
		return nil, fmt.Errorf("failed to create namespace")
	}

	container.Namespace = ns
	return container, nil
}

// AddNamespace assigns or updates the container's namespace
func (c *Container) AddNamespace(name string) error {
	if c.Namespace != nil {
		return fmt.Errorf("container already has a namespace: %s", c.Namespace.Name)
	}

	ns := CreateNamespace(name)
	if ns == nil {
		return fmt.Errorf("failed to create namespace")
	}

	c.Namespace = ns
	return nil
}

// AddBridge creates and adds a bridge to the container's namespace
func (c *Container) AddBridge(name string) (*Bridge, error) {
	if c.Namespace == nil {
		return nil, fmt.Errorf("container does not have a namespace")
	}

	bridge, err := CreateBridge(name, c.Namespace)
	if err != nil {
		return nil, fmt.Errorf("create bridge: %w", err)
	}

	c.Bridges = append(c.Bridges, *bridge)
	return bridge, nil
}

// AddVeth creates and adds a veth pair to the container
func (c *Container) AddVeth(nsA, nsB *Namespace, name string) (*Veth, error) {
	veth, err := CreateVeth(nsA, nsB, name)
	if err != nil {
		return nil, fmt.Errorf("create veth: %w", err)
	}

	c.Veths = append(c.Veths, *veth)
	return veth, nil
}

// ConnectVethToBridge connects a veth end to a bridge in the container's namespace
func (c *Container) ConnectVethToBridge(vethName, bridgeName, vethEnd string) error {
	if c.Namespace == nil {
		return fmt.Errorf("container does not have a namespace")
	}

	// Find the veth
	var veth *Veth
	for i := range c.Veths {
		if c.Veths[i].Name == vethName {
			veth = &c.Veths[i]
			break
		}
	}
	if veth == nil {
		return fmt.Errorf("veth %s not found", vethName)
	}

	// Find the bridge
	var bridge *Bridge
	for i := range c.Bridges {
		if c.Bridges[i].Name == bridgeName {
			bridge = &c.Bridges[i]
			break
		}
	}
	if bridge == nil {
		return fmt.Errorf("bridge %s not found", bridgeName)
	}

	// Move veth end to namespace
	if err := veth.MoveEndToNamespace(vethEnd, c.Namespace); err != nil {
		return fmt.Errorf("move veth: %w", err)
	}

	// Add interface to bridge
	if err := bridge.AddInterface(*veth); err != nil {
		return fmt.Errorf("add interface to bridge: %w", err)
	}

	return nil
}

// Exec executes a command inside the container's namespace
func (c *Container) Exec(cmd []string) error {
	if c.Namespace == nil {
		return fmt.Errorf("container does not have a namespace")
	}

	// Lock OS thread for namespace operations
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get current namespace
	origNS, err := os.Open("/proc/self/ns/net")
	if err != nil {
		return fmt.Errorf("get current namespace: %w", err)
	}
	defer origNS.Close()

	// Open target namespace
	targetNS, err := os.Open(c.Namespace.Path)
	if err != nil {
		return fmt.Errorf("open namespace: %w", err)
	}
	defer targetNS.Close()

	// Enter the container namespace
	if err := unix.Setns(int(targetNS.Fd()), unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Restore original namespace before returning
	defer unix.Setns(int(origNS.Fd()), unix.CLONE_NEWNET)

	// Run the command in the namespace
	execution := exec.Command(cmd[0], cmd[1:]...)
	execution.Stdout = os.Stdout
	execution.Stderr = os.Stderr
	execution.Stdin = os.Stdin

	return execution.Run()
}

// AttachShell attaches to an interactive shell in the container's namespace
func (c *Container) AttachShell() error {
	if c.Namespace == nil {
		return fmt.Errorf("container does not have a namespace")
	}

	// Check if we're being called as a child process
	if len(os.Args) >= 2 && os.Args[1] == "__gonett_nsenter__" {
		return c.nsenterChild()
	}

	// Parent process: fork and exec ourselves with special flag
	cmd := exec.Command("/proc/self/exe", "__gonett_nsenter__", c.Namespace.Path, c.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// nsenterChild is executed in the child process to enter the namespace
func (c *Container) nsenterChild() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("missing arguments for nsenter")
	}

	nsPath := os.Args[2]
	containerName := os.Args[3]

	// Lock the OS thread to ensure namespace operations affect this thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Open the network namespace file
	nsFd, err := os.Open(nsPath)
	if err != nil {
		return fmt.Errorf("open namespace: %w", err)
	}
	defer nsFd.Close()

	// Enter the network namespace using setns syscall
	if err := unix.Setns(int(nsFd.Fd()), unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("setns: %w", err)
	}

	// Set up environment
	ps1 := fmt.Sprintf("gonett@%s:\\w $ ", containerName)
	os.Setenv("PS1", ps1)

	// Execute bash shell
	bashArgs := []string{
		"bash",
		"--noprofile",
		"--norc",
	}

	if err := syscall.Exec("/bin/bash", bashArgs, os.Environ()); err != nil {
		return fmt.Errorf("exec bash: %w", err)
	}

	return nil
}

func attachToChild(name string) {
	if len(os.Args) < 3 {
		fmt.Println("child: missing netns path")
		os.Exit(1)
	}

	nsPath := filepath.Join(NETNS_BASE, name)

	fd, err := os.Open(nsPath)
	if err != nil {
		fmt.Println("child: open:", err)
		os.Exit(1)
	}
	defer fd.Close()

	if err := unix.Setns(int(fd.Fd()), unix.CLONE_NEWNET); err != nil {
		fmt.Println("child: setns:", err)
		os.Exit(1)
	}

	unix.Sethostname([]byte(name))
	ps1 := `gonett@\h:\w $ `

	cmd := []string{
		"bash",
		"-c",
		fmt.Sprintf("export PS1='%s'; exec bash --noprofile --norc", ps1),
	}

	if err := syscall.Exec("/bin/bash", cmd, os.Environ()); err != nil {
		fmt.Println("child: exec:", err)
		os.Exit(1)
	}
}

func (c *Container) Attach() error {
	if c.isChild {
		attachToChild(c.Name)
	}

	cmd := exec.Command("/proc/self/exe", "child", c.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(os.Environ(), "GONETT_CHILD_AUTH=true")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty:    true,
		Setsid:     true,
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	return cmd.Run()
}
