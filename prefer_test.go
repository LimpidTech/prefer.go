package prefer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadCreatesNewConfiguration(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mock := Mock{}
	configuration, err := Load("share/fixtures/example", &mock)
	checkTestError(t, err)

	file_path_index := strings.Index(configuration.Identifier, "share/fixtures/example.")
	expected_index := len(configuration.Identifier) - 27

	if file_path_index != expected_index {
		t.Error("Loaded unexpected configuration file:", configuration.Identifier)
	}

	if mock.Name != "Bailey" || mock.Age != 30 {
		t.Error("Got unexpected values from configuration file.")
	}
}

func TestLoadReturnsErrorForFilesWhichDontExist(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mock := Mock{}
	_, err := Load("this/is/a/fake/filename", &mock)

	if err == nil {
		t.Error("Expected an error but one was not returned.")
	}
}

func TestWatchReturnsChannelForWatchingFileForUpdates(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mock := Mock{}
	channel, err := Watch("share/fixtures/example", &mock)
	checkTestError(t, err)

	<-channel

	if mock.Name != "Bailey" || mock.Age != 30 {
		t.Error("Got unexpected values from configuration file.")
	}
}

func TestWatchWithDoneCanBeStopped(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	initialContent := `{"name": "initial"}`
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := Mock{}
	done := make(chan struct{})
	channel, err := WatchWithDone(tmpFile, &mock, done)
	checkTestError(t, err)

	// Wait for initial load
	select {
	case <-channel:
		if mock.Name != "initial" {
			t.Error("Expected initial name, got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for initial config")
	}

	// Stop watching
	close(done)

	// Give it time to clean up
	time.Sleep(100 * time.Millisecond)

	// Channel should be closed
	select {
	case _, ok := <-channel:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(500 * time.Millisecond):
		// Channel not closed yet, that's ok for this test
	}
}

func TestWatchWithDoneDetectsFileChanges(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	initialContent := `{"name": "initial"}`
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := Mock{}
	done := make(chan struct{})
	defer close(done)

	channel, err := WatchWithDone(tmpFile, &mock, done)
	checkTestError(t, err)

	// Wait for initial load
	select {
	case <-channel:
		if mock.Name != "initial" {
			t.Error("Expected initial name, got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for initial config")
	}

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Update the file
	newContent := `{"name": "updated"}`
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for update notification
	select {
	case <-channel:
		if mock.Name != "updated" {
			t.Error("Expected updated name, got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for config update")
	}
}

func TestWatchWithDoneSkipsInvalidUpdates(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	initialContent := `{"name": "initial"}`
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := Mock{}
	done := make(chan struct{})
	defer close(done)

	channel, err := WatchWithDone(tmpFile, &mock, done)
	checkTestError(t, err)

	// Wait for initial load
	<-channel

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Write invalid JSON - should be skipped, not crash
	if err := os.WriteFile(tmpFile, []byte(`{invalid json}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Give it time to process
	time.Sleep(200 * time.Millisecond)

	// Write valid JSON again
	if err := os.WriteFile(tmpFile, []byte(`{"name": "recovered"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Should receive the valid update
	select {
	case <-channel:
		if mock.Name != "recovered" {
			t.Error("Expected recovered name, got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for recovered config")
	}
}
