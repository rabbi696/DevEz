// Package paths resolves where DevEz keeps its config, data and sockets on each OS.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Root returns DevEz's data directory: ~/Library/Application Support/DevEz on macOS,
// %AppData%\DevEz on Windows, $XDG_DATA_HOME/devez (~/.local/share/devez) on Linux.
// DEVEZ_HOME (or legacy FLOCK_HOME) overrides it (used by tests, for running isolated instances,
// and by system services).
func Root() (string, error) {
	if dir := os.Getenv("DEVEZ_HOME"); dir != "" {
		return dir, nil
	}
	if dir := os.Getenv("FLOCK_HOME"); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "linux" {
		return dataHome()
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, "DevEz"), nil
}

// dataHome is $XDG_DATA_HOME/devez, ~/.local/share/devez when it is unset or relative (which
// the XDG spec says to ignore).
func dataHome() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "devez"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".local", "share", "devez"), nil
}

// Socket is the daemon's API socket. AF_UNIX works on macOS, Linux and Windows 10+.
func Socket(root string) string {
	return filepath.Join(root, "devezd.sock")
}

// Run is the directory for the services' unix sockets (PHP-FPM, MySQL/MariaDB, Postgres). Kept
// short: AF_UNIX paths are limited to ~104 bytes on macOS.
func Run(root string) string {
	return filepath.Join(root, "run")
}

// EnsureRun creates Run(root) and returns it. Every process that puts a socket there calls it
// before starting: nothing else creates the directory.
func EnsureRun(root string) (string, error) {
	dir := Run(root)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
