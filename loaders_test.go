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
