package prefer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestFileLoaderWatchWithContextDetectsFileChanges(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	initialContent := `{"name": "initial"}`
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	if err := loader.WatchWithContext(channel, done); err != nil {
		t.Fatal(err)
	}

	// Give the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	newContent := `{"name": "updated"}`
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for notification
	select {
	case <-channel:
		// Success - received update notification
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for file change notification")
	}

	// Cleanup
	close(done)
}

func TestFileLoaderWatchWithContextStopsOnDone(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	if err := loader.WatchWithContext(channel, done); err != nil {
		t.Fatal(err)
	}

	// Give the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Close done channel to stop watching
	close(done)

	// Give it time to clean up
	time.Sleep(100 * time.Millisecond)

	// Write to file - should not receive notification since watcher stopped
	if err := os.WriteFile(tmpFile, []byte(`{"updated": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-channel:
		t.Error("Received notification after done channel was closed")
	case <-time.After(200 * time.Millisecond):
		// Success - no notification received
	}
}

func TestFileLoaderWatchWithContextReturnsErrorForNonexistentFile(t *testing.T) {
	loader := FileLoader{identifier: "nonexistent/file.json"}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	err := loader.WatchWithContext(channel, done)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestFileLoaderWatchBackwardsCompatibility(t *testing.T) {
	// Test that the old Watch method still works
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	// This should not block or return error
	err := loader.Watch(channel)
	if err != nil {
		t.Error("Watch returned unexpected error:", err)
	}

	// Give the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(tmpFile, []byte(`{"updated": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for notification
	select {
	case <-channel:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for file change notification")
	}
}

func TestFileLoaderLocateWithAbsolutePath(t *testing.T) {
	// Create a temporary file with absolute path
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	loader := FileLoader{identifier: tmpFile}
	location, err := loader.Locate()
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	if location != tmpFile {
		t.Error("Expected location to be", tmpFile, "got", location)
	}
}

func TestFileLoaderLocateWithAbsolutePathWithoutExtension(t *testing.T) {
	// Create a temporary file - test finding it via extension detection
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Search with absolute path but without extension
	identifier := filepath.Join(tmpDir, "config")
	loader := FileLoader{identifier: identifier}
	location, err := loader.Locate()
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	if location != tmpFile {
		t.Error("Expected location to be", tmpFile, "got", location)
	}
}

func TestFileLoaderLocateWithAbsolutePathNotFound(t *testing.T) {
	// Test with absolute path that doesn't exist
	loader := FileLoader{identifier: "/nonexistent/absolute/path/config"}
	_, err := loader.Locate()
	if err == nil {
		t.Error("Expected error for non-existent absolute path")
	}
}

func TestFileLoaderLoadWithExtensionLoop(t *testing.T) {
	// Create multiple files with different extensions
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(yamlFile, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Use identifier without extension - should find the yaml file
	identifier := filepath.Join(tmpDir, "config")
	loader := FileLoader{identifier: identifier}

	location, content, err := loader.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if location != yamlFile {
		t.Error("Expected location to be", yamlFile, "got", location)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty content")
	}
}

func TestFileLoaderLoadUnreadableFile(t *testing.T) {
	// Skip on CI where permission tests may not work
	if os.Getenv("CI") != "" {
		t.Skip("Skipping permission test in CI")
	}

	// Create a file without read permissions
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(tmpFile, 0644) // Restore permissions for cleanup

	loader := FileLoader{identifier: tmpFile}
	_, _, err := loader.Load()
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
}

func TestFileLoaderWatchWithContextNilDone(t *testing.T) {
	// Test watching without a done channel (nil done)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	// Watch with nil done channel
	err := loader.WatchWithContext(channel, nil)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	// Give the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(tmpFile, []byte(`{"updated": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for notification
	select {
	case <-channel:
		// Success - received update notification
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for file change notification")
	}
}
