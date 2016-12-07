package prefer

import (
	"bufio"
	"errors"
	"os"
	"path"
)

type Loader interface {
	Load(identifier string) (string, []byte, error)
}

func NewLoader(identifier string) (Loader, error) {
	switch identifier {
	default:
		return FileLoader{}, nil
	}
}

type FileLoader struct{}

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

func (loader FileLoader) Locate(identifier string) (string, error) {
	for index := range standardPaths {
		directory := standardPaths[index]
		identifierWithPath := path.Join(directory, identifier)

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

func (loader FileLoader) Load(identifier string) (string, []byte, error) {
	location, err := loader.Locate(identifier)

	if err != nil {
		return "", nil, err
	}

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
