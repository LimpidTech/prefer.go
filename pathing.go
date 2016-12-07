package prefer

import (
	"log"
	"os"
	"path/filepath"
)

var standardPaths []string

func GetStandardPaths() []string {
	paths := make([]string, len(standardPaths))
	copy(paths, standardPaths)
	return paths
}

func init() {
	pathMap := make(map[string]interface{})
	wd, err := os.Getwd()

	if err != nil {
		log.Fatalln("Could not get current working directory.")
	}

	// Remove /bin if it's at the end of the cwd
	// NOTE: Er, os.PathSeparator is a rune... So, here's a hack.
	if len(wd) > 4 && wd[len(wd)-4:] == filepath.Join("", "bin") {
		wd = wd[:len(wd)-4]
	}

	paths := []string{".", wd}
	xdgPaths := filepath.SplitList(os.Getenv("XDG_CONFIG_DIRS"))

	if len(xdgPaths) > 0 {
		paths = append(paths, xdgPaths...)
	}

	paths = append(paths, getSystemPaths()...)

	for _, path := range paths {
		if path == "/" {
			path = ""
		}

		pathMap[filepath.Join(path, "/etc")] = nil

		if len(path) > 0 {
			pathMap[path] = nil
		}
	}

	for path := range pathMap {
		standardPaths = append(standardPaths, path)
	}
}
