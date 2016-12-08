package prefer

import (
	"strings"
	"testing"
)

func TestFileLoaderLocatesFilesWithExtensions(t *testing.T) {
	pathName := "share/fixtures/example.yaml"
	loader, err := NewLoader(pathName)
	checkTestError(t, err)

	identifier, _, err := loader.Load()
	if strings.HasSuffix(identifier, pathName) != true {
		t.Error("Unexpected result from Load()")
	}
}

func TestFileLoaderLocatesFilesWithoutExtensions(t *testing.T) {
	pathName := "share/fixtures/example"
	loader, err := NewLoader(pathName)
	checkTestError(t, err)

	identifier, _, err := loader.Load()
	if strings.HasSuffix(identifier, pathName+".yaml") != true {
		t.Error("Unexpected result from Load()")
	}
}
