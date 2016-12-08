package prefer

import (
	"strings"
	"testing"
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
