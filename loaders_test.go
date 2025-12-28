package prefer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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

func TestNewLoaderWithEmptyIdentifier(t *testing.T) {
	_, err := NewLoader("")
	if err == nil {
		t.Error("Expected error for empty identifier")
	}
}

func TestCheckFileExistsWithStatError(t *testing.T) {
	// Save original statFunc and restore after test
	originalStatFunc := statFunc
	defer func() { statFunc = originalStatFunc }()

	// Mock statFunc to return a non-IsNotExist error
	statFunc = func(name string) (os.FileInfo, error) {
		return nil, errors.New("permission denied")
	}

	exists, err := checkFileExists("/some/path")
	if !exists {
		t.Error("Expected exists to be true when stat returns non-IsNotExist error")
	}
	if err == nil {
		t.Error("Expected error to be returned")
	}
	if err.Error() != "permission denied" {
		t.Error("Expected 'permission denied' error, got:", err.Error())
	}
}

// mockWatcher implements the Watcher interface for testing
type mockWatcher struct {
	events     chan fsnotify.Event
	errors     chan error
	addErr     error
	closeErr   error
	closeCalls atomic.Int32
}

func newMockWatcher() *mockWatcher {
	return &mockWatcher{
		events: make(chan fsnotify.Event, 10),
		errors: make(chan error, 10),
	}
}

func (m *mockWatcher) Add(name string) error { return m.addErr }
func (m *mockWatcher) Close() error {
	m.closeCalls.Add(1)
	return m.closeErr
}
func (m *mockWatcher) Events() <-chan fsnotify.Event { return m.events }
func (m *mockWatcher) Errors() <-chan error          { return m.errors }

func TestWatchWithContextEventsChannelClosed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	err := loader.WatchWithContext(channel, done)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Close the events channel to trigger the !ok branch
	close(mock.events)

	// Give the goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	// The watcher should have been closed
	if mock.closeCalls.Load() == 0 {
		t.Error("Expected watcher to be closed")
	}
}

func TestWatchWithContextErrorsChannelClosed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	err := loader.WatchWithContext(channel, done)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Close the errors channel to trigger the !ok branch
	close(mock.errors)

	// Give the goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	// The watcher should have been closed
	if mock.closeCalls.Load() == 0 {
		t.Error("Expected watcher to be closed")
	}
}

func TestWatchWithContextReceivesError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)
	done := make(chan struct{})

	err := loader.WatchWithContext(channel, done)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Send an error - should be ignored (resilient watching)
	mock.errors <- errors.New("some error")

	// Give it time to process
	time.Sleep(50 * time.Millisecond)

	// Send a valid event - should still work
	mock.events <- fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	select {
	case <-channel:
		// Success - watcher continued after error
	case <-time.After(500 * time.Millisecond):
		t.Error("Watcher should continue after error")
	}

	close(done)
}

func TestWatchWithContextNewWatcherError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	newWatcher = func() (Watcher, error) {
		return nil, errors.New("failed to create watcher")
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	err := loader.WatchWithContext(channel, nil)
	if err == nil {
		t.Error("Expected error when newWatcher fails")
	}
}

func TestWatchWithContextAddError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	mock.addErr = errors.New("failed to add watch")
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	err := loader.WatchWithContext(channel, nil)
	if err == nil {
		t.Error("Expected error when Add fails")
	}
	if mock.closeCalls.Load() == 0 {
		t.Error("Expected watcher to be closed on Add error")
	}
}

func TestWatchWithContextNilDoneEventsChannelClosed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	// Watch with nil done channel
	err := loader.WatchWithContext(channel, nil)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Close the events channel to trigger the !ok branch in the nil done path
	close(mock.events)

	// Give the goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	if mock.closeCalls.Load() == 0 {
		t.Error("Expected watcher to be closed")
	}
}

func TestWatchWithContextNilDoneErrorsChannelClosed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	// Watch with nil done channel
	err := loader.WatchWithContext(channel, nil)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Close the errors channel to trigger the !ok branch in the nil done path
	close(mock.errors)

	// Give the goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	if mock.closeCalls.Load() == 0 {
		t.Error("Expected watcher to be closed")
	}
}

func TestWatchWithContextNilDoneReceivesError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original newWatcher and restore after test
	originalNewWatcher := newWatcher
	defer func() { newWatcher = originalNewWatcher }()

	mock := newMockWatcher()
	newWatcher = func() (Watcher, error) {
		return mock, nil
	}

	loader := FileLoader{identifier: tmpFile}
	channel := make(chan bool, 1)

	// Watch with nil done channel
	err := loader.WatchWithContext(channel, nil)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Send an error - should be ignored (resilient watching)
	mock.errors <- errors.New("some error")

	// Give it time to process
	time.Sleep(50 * time.Millisecond)

	// Send a valid event - should still work
	mock.events <- fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	select {
	case <-channel:
		// Success - watcher continued after error
	case <-time.After(500 * time.Millisecond):
		t.Error("Watcher should continue after error")
	}
}

func TestMemoryLoaderLoad(t *testing.T) {
	content := []byte(`{"name": "test"}`)
	loader := NewMemoryLoader("config.json", content)

	identifier, data, err := loader.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if identifier != "config.json" {
		t.Error("Expected identifier 'config.json', got:", identifier)
	}

	if string(data) != string(content) {
		t.Error("Content mismatch")
	}
}

func TestMemoryLoaderWatchReturnsError(t *testing.T) {
	loader := NewMemoryLoader("config.json", []byte(`{}`))

	channel := make(chan bool)
	err := loader.Watch(channel)
	if err == nil {
		t.Error("Expected error from MemoryLoader.Watch")
	}
}

func TestMemoryLoaderWatchWithContextReturnsError(t *testing.T) {
	loader := NewMemoryLoader("config.json", []byte(`{}`))

	channel := make(chan bool)
	done := make(chan struct{})
	err := loader.WatchWithContext(channel, done)
	if err == nil {
		t.Error("Expected error from MemoryLoader.WatchWithContext")
	}
}
