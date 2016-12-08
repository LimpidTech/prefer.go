package prefer

import (
	"os"
	"runtime"
	"testing"
)

func TestGetStandardPaths(t *testing.T) {
	os.Setenv("XDG_CONFIG_DIRS", "")

	switch runtime.GOOS {
	case "windows":
		t.Error("You are using a poorly designed operating system.")

	default:
		if len(GetStandardPaths()) != 13 {
			t.Error("Got unexpected number of paths from GetStandardPaths().")
		}
	}
}
