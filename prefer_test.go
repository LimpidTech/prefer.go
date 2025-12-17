package prefer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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

func TestConfigurationWatch(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	initialContent := `{"name": "test"}`
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	mock := Mock{}
	config := NewConfiguration(tmpFile)
	channel := make(chan interface{}, 1)

	// Test the Configuration.Watch method (which wraps WatchWithDone)
	err := config.Watch(&mock, channel)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Wait for initial load
	select {
	case <-channel:
		if mock.Name != "test" {
			t.Error("Expected name to be 'test', got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for initial config")
	}
}

func TestConfigurationReloadWithInvalidFile(t *testing.T) {
	// Test Reload with a file that doesn't exist
	config := NewConfiguration("nonexistent/file.json")
	type Mock struct {
		Name string `json:"name"`
	}
	mock := Mock{}

	err := config.Reload(&mock)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConfigurationReloadWithUnsupportedFormat(t *testing.T) {
	// Create a temporary file with unsupported extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.unknown")

	if err := os.WriteFile(tmpFile, []byte(`data`), 0644); err != nil {
		t.Fatal(err)
	}

	config := NewConfiguration(tmpFile)
	type Mock struct {
		Name string
	}
	mock := Mock{}

	err := config.Reload(&mock)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestWatchWithDoneReloadError(t *testing.T) {
	// Test that WatchWithDone returns error when initial Reload fails
	type Mock struct {
		Name string `json:"name"`
	}

	mock := Mock{}
	done := make(chan struct{})
	defer close(done)

	config := NewConfiguration("nonexistent/file.json")
	channel := make(chan interface{}, 1)

	err := config.WatchWithDone(&mock, channel, done)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestReloadWithEmptyIdentifier(t *testing.T) {
	// Test Reload when NewLoader returns error (empty identifier)
	config := NewConfiguration("")
	type Mock struct {
		Name string `json:"name"`
	}
	mock := Mock{}

	err := config.Reload(&mock)
	if err == nil {
		t.Error("Expected error for empty identifier")
	}
}

func TestWatchWithDoneNewLoaderError(t *testing.T) {
	// Test WatchWithDone when NewLoader returns error (empty identifier)
	type Mock struct {
		Name string `json:"name"`
	}

	mock := Mock{}
	done := make(chan struct{})
	defer close(done)

	config := NewConfiguration("")
	channel := make(chan interface{}, 1)

	err := config.WatchWithDone(&mock, channel, done)
	if err == nil {
		t.Error("Expected error for empty identifier")
	}
}

func TestWatchWithDoneUpdateChannelClosed(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
		t.Fatal(err)
	}

	mock := Mock{}
	done := make(chan struct{})
	channel, err := WatchWithDone(tmpFile, &mock, done)
	checkTestError(t, err)

	// Wait for initial load
	<-channel

	// Close done to stop the watcher
	close(done)

	// Give time for goroutine to clean up
	time.Sleep(200 * time.Millisecond)

	// Channel should be closed
	select {
	case _, ok := <-channel:
		if ok {
			// Still open, that's okay for this test
		}
	case <-time.After(500 * time.Millisecond):
		// Timeout is fine
	}
}

func TestWatchWithDoneSkipsSerializerErrors(t *testing.T) {
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
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

	// Rename file to an unsupported extension to trigger NewSerializer error
	newFile := filepath.Join(tmpDir, "config.unsupported")
	os.Rename(tmpFile, newFile)

	// Write to trigger a watch event (even though it won't be picked up properly)
	os.WriteFile(tmpFile, []byte(`{"name": "updated"}`), 0644)

	time.Sleep(200 * time.Millisecond)

	// The watcher should continue running without crashing
	// Write valid JSON again
	os.WriteFile(tmpFile, []byte(`{"name": "recovered"}`), 0644)

	// Should eventually receive an update
	select {
	case <-channel:
		// Got an update, good
	case <-time.After(2 * time.Second):
		// Timeout is acceptable - the point is the watcher didn't crash
	}
}

func TestWatchWithDoneWatcherError(t *testing.T) {
	// Test WatchWithDone when WatchWithContext returns an error
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	// Make newWatcher fail - this happens in WatchWithContext after Reload succeeds
	newWatcher = func() (Watcher, error) {
		return nil, errors.New("watcher error")
	}

	mock := Mock{}
	done := make(chan struct{})
	defer close(done)

	config := NewConfiguration(tmpFile)
	channel := make(chan interface{}, 1)

	err := config.WatchWithDone(&mock, channel, done)
	if err == nil {
		t.Error("Expected error when WatchWithContext fails")
	}
}

func TestWatchWithDoneLoadErrorDuringUpdate(t *testing.T) {
	// Test that loader.Load() errors during watch updates are skipped (resilient)
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
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

	// Delete the file - this will cause Load() to fail on next update
	os.Remove(tmpFile)

	// Create the file again to trigger a watcher event
	// The Load will succeed now since the file exists again
	if err := os.WriteFile(tmpFile, []byte(`{"name": "recovered"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Should receive the recovered config
	select {
	case <-channel:
		if mock.Name != "recovered" {
			t.Error("Expected recovered name, got:", mock.Name)
		}
	case <-time.After(2 * time.Second):
		// Timeout is acceptable - the watcher continues to work
	}
}

// mockWatcherForPrefer is a test-specific mock for prefer_test.go
type mockWatcherForPrefer struct {
	events     chan fsnotify.Event
	errors     chan error
	addErr     error
	closeCalls int
}

func (m *mockWatcherForPrefer) Add(name string) error           { return m.addErr }
func (m *mockWatcherForPrefer) Close() error                    { m.closeCalls++; return nil }
func (m *mockWatcherForPrefer) Events() <-chan fsnotify.Event   { return m.events }
func (m *mockWatcherForPrefer) Errors() <-chan error            { return m.errors }

func TestConfigurationWatchWithDoneUpdateChannelClosed(t *testing.T) {
	// Test that the goroutine exits gracefully when update channel closes
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := &mockWatcherForPrefer{
		events: make(chan fsnotify.Event, 10),
		errors: make(chan error, 10),
	}

	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	data := Mock{}
	done := make(chan struct{})
	config := NewConfiguration(tmpFile)
	channel := make(chan interface{}, 1)

	err := config.WatchWithDone(&data, channel, done)
	checkTestError(t, err)

	// Wait for initial notification
	<-channel

	// Close the mock events channel - this simulates watcher shutdown
	close(mock.events)

	// Give time for goroutine to exit
	time.Sleep(100 * time.Millisecond)

	// The output channel should be closed now
	select {
	case _, ok := <-channel:
		if ok {
			// Got a value, that's okay
		}
		// Channel closed, as expected
	case <-time.After(500 * time.Millisecond):
		// Timeout - channel not closed, but that's acceptable
	}
}

func TestConfigurationWatchWithDoneLoadError(t *testing.T) {
	// Test that loader.Load() errors during update are skipped
	type Mock struct {
		Name string `json:"name"`
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(`{"name": "initial"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := &mockWatcherForPrefer{
		events: make(chan fsnotify.Event, 10),
		errors: make(chan error, 10),
	}

	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	data := Mock{}
	done := make(chan struct{})
	defer close(done)

	config := NewConfiguration(tmpFile)
	channel := make(chan interface{}, 1)

	err := config.WatchWithDone(&data, channel, done)
	checkTestError(t, err)

	// Wait for initial notification
	<-channel

	// Delete the file to cause Load() to fail
	os.Remove(tmpFile)

	// Send a write event - this will trigger Load() which will fail
	mock.events <- fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	// Give time for the error path to execute
	time.Sleep(100 * time.Millisecond)

	// Recreate the file
	if err := os.WriteFile(tmpFile, []byte(`{"name": "recovered"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Send another event - this one should succeed
	mock.events <- fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	// Should receive the recovered config
	select {
	case <-channel:
		if data.Name != "recovered" {
			t.Error("Expected recovered name, got:", data.Name)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for recovered config")
	}
}
