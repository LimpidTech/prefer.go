package prefer

import (
	"log"
	"os"
	"path/filepath"
)

var standardPaths []string

func StandardPaths() []string {
	paths := make([]string, len(standardPaths))
	copy(paths, standardPaths)
	return paths
}

func distinct(items []string) (result []string) {
Search:
	for _, item := range items {
		for _, existing := range result {
			if existing == item {
				continue Search
			}
		}

		result = append(result, item)
	}

	return result
}

func init() {
	wd, err := os.Getwd()

	if err != nil {
		log.Fatalln("Could not get current working directory.")
	}

	// Remove /bin if it's at the end of the cwd
	// TODO: Er, os.PathSeparator is a rune... So, here's a hack.
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

		standardPaths = append(standardPaths, filepath.Join(path, "/etc"))

		if path == "" {
			continue
		}

		if len(path) > 0 {
			standardPaths = append(standardPaths, path)
		}
	}

	standardPaths = distinct(standardPaths)
}
