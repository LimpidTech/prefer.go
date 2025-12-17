package prefer

import (
	"runtime"
	"testing"
)

func TestGetStandardPaths(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Error("You are using a poorly designed operating system.")

	default:
		// The number of paths depends on the environment (e.g., XDG_CONFIG_DIRS)
		// We expect at least 10 paths: ., cwd, /etc, and various system paths
		paths := GetStandardPaths()
		if len(paths) < 10 {
			t.Error("Got unexpected number of paths from GetStandardPaths():", len(paths))
		}
	}
}

func TestGetStandardPathsReturnsCopy(t *testing.T) {
	// Verify that GetStandardPaths returns a copy, not the original slice
	paths1 := GetStandardPaths()
	paths2 := GetStandardPaths()

	if len(paths1) == 0 {
		t.Fatal("Expected non-empty paths")
	}

	// Modify paths1
	paths1[0] = "modified"

	// paths2 should be unaffected
	if paths2[0] == "modified" {
		t.Error("GetStandardPaths should return a copy, not the original")
	}
}
