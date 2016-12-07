package prefer

import (
	"reflect"
	"testing"
)

const (
	MOCK_NAME  = "Mock Name"
	MOCK_VALUE = 12
)

type MockSubject struct {
	Name  string
	Value int
}

func checkTestError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func getMockSubject() *MockSubject {
	return &MockSubject{
		Name:  MOCK_NAME,
		Value: MOCK_VALUE,
	}
}

func getMockSubjectSerialize(t *testing.T, serializer Serializer) []byte {
	subject := getMockSubject()

	serialized, err := serializer.Serialize(subject)
	checkTestError(t, err)

	return serialized
}

func TestNewSerializerReturnsYAMLSerializer(t *testing.T) {
	content := getMockSubjectSerialize(t, YAMLSerializer{})
	serializer, err := NewSerializer("example.yaml", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "YAMLSerializer" {
		t.Error("Got Serializer of wrong type when requesting YAMLSerializer.")
	}
}

func TestNewSerializerReturnsXMLSerializer(t *testing.T) {
	content := getMockSubjectSerialize(t, XMLSerializer{})
	serializer, err := NewSerializer("example.xml", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "XMLSerializer" {
		t.Error("Got Serializer of wrong type when requesting XMLSerializer.")
	}
}

func YAMLSerializerTestCase(t *testing.T) {
	serializer := YAMLSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := &MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}

func XMLSerializerTestCase(t *testing.T) {
	serializer := XMLSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := &MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}
