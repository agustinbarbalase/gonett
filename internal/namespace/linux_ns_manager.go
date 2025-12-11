package namespace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

const NETNS_DIR = "/var/run/netns"

type LinuxNamespaceManager struct{}

func NewLinuxNamespaceManager() NamespaceManager {
	return &LinuxNamespaceManager{}
}

func (ln *LinuxNamespaceManager) Create(name string) error {
	if err := os.MkdirAll(NETNS_DIR, 0755); err != nil {
		return err
	}

	nsPath := filepath.Join(NETNS_DIR, name)

	f, err := os.Create(nsPath)
	if err != nil {
		return err
	}
	f.Close()

	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("unshare: %w", err)
	}

	if err := unix.Mount("/proc/self/ns/net", nsPath, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("bind mount: %w", err)
	}

	return nil
}

func (ln *LinuxNamespaceManager) Delete(name string) error {
	nsPath := filepath.Join(NETNS_DIR, name)

	if err := unix.Unmount(nsPath, 0); err != nil {
		return fmt.Errorf("umount: %w", err)
	}

	if err := os.Remove(nsPath); err != nil {
		return fmt.Errorf("rm: %w", err)
	}

	return nil
}

func (ln *LinuxNamespaceManager) Execute(name string, command []string) error {
	cmd := exec.Command("/proc/self/exe", "child", name)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}

	return cmd.Run()
}

func (ln *LinuxNamespaceManager) List() ([]string, error) {
	entries, err := os.ReadDir(NETNS_DIR)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, entry := range entries {
		namespaces = append(namespaces, entry.Name())
	}

	return namespaces, nil
}
