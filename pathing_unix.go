//go:build !windows

package prefer

import (
	"os"
	"path/filepath"
)

func getSystemPaths() []string {
	return []string{
		filepath.Join(os.Getenv("HOME"), ".config"),
		os.Getenv("HOME"),
		"/usr/local",
		"/usr",
		"/",
	}
}
