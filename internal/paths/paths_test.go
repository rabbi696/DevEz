package paths

import (
	"runtime"
	"testing"
)

func TestDataHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG paths are Linux-only")
	}
	t.Setenv("HOME", "/home/ada")

	t.Setenv("XDG_DATA_HOME", "")
	if got, _ := dataHome(); got != "/home/ada/.local/share/devez" {
		t.Errorf("unset: %q", got)
	}
	t.Setenv("XDG_DATA_HOME", "/data")
	if got, _ := dataHome(); got != "/data/devez" {
		t.Errorf("set: %q", got)
	}
	t.Setenv("XDG_DATA_HOME", "relative")
	if got, _ := dataHome(); got != "/home/ada/.local/share/devez" {
		t.Errorf("relative: %q", got)
	}
}

func TestRootPrefersFlockHome(t *testing.T) {
	t.Setenv("DEVEZ_HOME", "/tmp/devez")
	if got, _ := Root(); got != "/tmp/devez" {
		t.Errorf("Root DEVEZ_HOME: %q", got)
	}
	t.Setenv("DEVEZ_HOME", "")
	t.Setenv("FLOCK_HOME", "/tmp/flock")
	if got, _ := Root(); got != "/tmp/flock" {
		t.Errorf("Root FLOCK_HOME: %q", got)
	}
}
