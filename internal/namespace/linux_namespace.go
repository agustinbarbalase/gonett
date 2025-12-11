package namespace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	mgr "gonet/internal/namespace/manager"

	"golang.org/x/sys/unix"
)

type LinuxNamespace struct {
	name    string
	isChild bool
}

func attachToChild(name string) {
	if len(os.Args) < 3 {
		fmt.Println("child: missing netns path")
		os.Exit(1)
	}

	nsPath := filepath.Join(mgr.NETNS_BASE, name)

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

func NewLinuxNamespace(name string, isChild bool) Namespace {
	return LinuxNamespace{name: name, isChild: isChild}
}

func (ns LinuxNamespace) Attach() error {
	if ns.isChild {
		attachToChild(ns.name)
	}

	cmd := exec.Command("/proc/self/exe", "child", ns.name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty:    true,
		Setsid:     true,
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	return cmd.Run()
}

func (ns LinuxNamespace) Exec(cmd []string) error {
	// Implementation for executing a command in the namespace
	return nil
}
