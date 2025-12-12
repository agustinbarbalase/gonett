package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

func cmdNsenter() {
	if len(os.Args) < 4 {
		fmt.Println("Internal error: missing arguments for nsenter")
		os.Exit(1)
	}

	nsPath := os.Args[2]
	containerName := os.Args[3]

	// Lock the OS thread to ensure namespace operations affect this thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Open the network namespace file
	nsFd, err := os.Open(nsPath)
	if err != nil {
		fmt.Printf("Error opening namespace: %v\n", err)
		os.Exit(1)
	}
	defer nsFd.Close()

	// Enter the network namespace using setns syscall
	if err := unix.Setns(int(nsFd.Fd()), unix.CLONE_NEWNET); err != nil {
		fmt.Printf("Error entering namespace: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Error executing bash: %v\n", err)
		os.Exit(1)
	}
}
