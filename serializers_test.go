package prefer

import (
	"reflect"
	"testing"
)

const (
	MOCK_NAME  = "Mock Name"
	MOCK_VALUE = 30
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

func getMockSubject() MockSubject {
	return MockSubject{
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

func TestNewSerializerReturnsINISerializer(t *testing.T) {
	content := getMockSubjectSerialize(t, INISerializer{})
	serializer, err := NewSerializer("example.ini", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "INISerializer" {
		t.Error("Got Serializer of wrong type when requesting INISerializer.")
	}
}

func TestNewSerializerReturnsErrorForUnknownFormats(t *testing.T) {
	content := getMockSubjectSerialize(t, XMLSerializer{})
	_, err := NewSerializer("example.dat", content)

	if err == nil {
		t.Error("Expected error, but didn't get one.")
	}
}

func TestYAMLSerializer(t *testing.T) {
	serializer := YAMLSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, &result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}

func TestXMLSerializer(t *testing.T) {
	serializer := XMLSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, &result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}

func TestINISerializer(t *testing.T) {
	serializer := INISerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, &result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}

func TestNewSerializerReturnsJSONSerializer(t *testing.T) {
	content := getMockSubjectSerialize(t, JSONSerializer{})
	serializer, err := NewSerializer("example.json", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "JSONSerializer" {
		t.Error("Got Serializer of wrong type when requesting JSONSerializer.")
	}
}

func TestJSON5Serializer(t *testing.T) {
	serializer := JSONSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, &result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}
