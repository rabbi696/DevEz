package main

import (
	"os"
	"syscall"

	"flock/internal/cli"
)

func main() {
	// If called via legacy 'flock' name on Linux, handle potential util-linux flock pass-through.
	if path, ok := cli.PassThrough(os.Args); ok {
		err := syscall.Exec(path, os.Args, os.Environ())
		os.Stderr.WriteString("devez: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Shims (php, composer, ...) call this binary as `devez exec <tool>`.
	if tool := cli.ShimTool(os.Args[0]); tool != "" {
		os.Args = append([]string{os.Args[0], "exec", tool}, os.Args[1:]...)
	}

	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
