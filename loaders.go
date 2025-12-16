package prefer

import (
	"bufio"
	"errors"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

type Loader interface {
	Load() (string, []byte, error)
	Watch(channel chan bool) error
	WatchWithContext(channel chan bool, done <-chan struct{}) error
}

func NewLoader(identifier string) (Loader, error) {
	switch identifier {
	default:
		return FileLoader{
			identifier: identifier,
		}, nil
	}
}

type FileLoader struct {
	identifier string
}

func checkFileExists(location string) (bool, error) {
	_, err := os.Stat(location)

	if err == nil {
		return true, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func (this FileLoader) Locate() (string, error) {
	// Check if identifier is already an absolute path that exists
	if path.IsAbs(this.identifier) {
		if exists, err := checkFileExists(this.identifier); exists {
			return this.identifier, err
		}
		// Try with extensions
		for extension := range defaultSerializers {
			identifierWithExtension := this.identifier + extension
			if exists, err := checkFileExists(identifierWithExtension); exists {
				return identifierWithExtension, err
			}
		}
	}

	paths := GetStandardPaths()

	for index := range paths {
		directory := paths[index]
		identifierWithPath := path.Join(directory, this.identifier)

		if exists, err := checkFileExists(identifierWithPath); exists == true {
			return identifierWithPath, err
		}

		for extension := range defaultSerializers {
			identifierWithExtension := identifierWithPath + extension

			if exists, err := checkFileExists(identifierWithExtension); exists == true {
				return identifierWithExtension, err
			}
		}
	}

	return "", errors.New("Could not find a configuration in the given location.")
}

func (this FileLoader) Load() (string, []byte, error) {
	location, err := this.Locate()

	if err != nil {
		return "", nil, err
	}

	this.identifier = location

	file, err := os.Open(location)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	var result []byte
	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		result = append(result, scanner.Bytes()...)
	}

	return location, result, err
}

// WatchEvent represents a file system event during watching
type WatchEvent struct {
	Path  string
	Error error
}

func (this FileLoader) Watch(channel chan bool) error {
	return this.WatchWithContext(channel, nil)
}

// WatchWithContext watches for file changes with support for graceful shutdown.
// Close the done channel to stop watching.
func (this FileLoader) WatchWithContext(channel chan bool, done <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Locate the file first to get the full path
	location, err := this.Locate()
	if err != nil {
		watcher.Close()
		return err
	}

	if err = watcher.Add(location); err != nil {
		watcher.Close()
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			if done != nil {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// Only notify on write/create events, like JS and Rust implementations
					if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
						channel <- true
					}
				case _, ok := <-watcher.Errors:
					if !ok {
						return
					}
					// Continue watching on errors (resilient like Rust)
					continue
				case <-done:
					return
				}
			} else {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
						channel <- true
					}
				case _, ok := <-watcher.Errors:
					if !ok {
						return
					}
					continue
				}
			}
		}
	}()

	return nil
}
