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

func (this FileLoader) Watch(channel chan bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err = watcher.Add(this.identifier); err != nil {
		return err
	}

	for {
		select {
		case <-watcher.Events:
			channel <- true
		case <-watcher.Errors:
			continue
		}
	}
}
