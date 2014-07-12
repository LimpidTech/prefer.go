package prefer

import (
	"os"
	"runtime"
	"testing"
)

func TestStandardPaths(t *testing.T) {
	os.Setenv("XDG_CONFIG_DIRS", "")

	switch runtime.GOOS {
	case "windows":
		t.Error("You are using a poorly designed operating system.")

	default:
		if len(StandardPaths()) != 13 {
			t.Error("Got unexpected number of paths from StandardPaths().")
		}
	}
}
