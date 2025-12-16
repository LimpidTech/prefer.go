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

func TestNewSerializerReturnsTOMLSerializer(t *testing.T) {
	content := getMockSubjectSerialize(t, TOMLSerializer{})
	serializer, err := NewSerializer("example.toml", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "TOMLSerializer" {
		t.Error("Got Serializer of wrong type when requesting TOMLSerializer.")
	}
}

func TestTOMLSerializer(t *testing.T) {
	serializer := TOMLSerializer{}
	serialized := getMockSubjectSerialize(t, serializer)

	result := MockSubject{}
	checkTestError(t, serializer.Deserialize(serialized, &result))

	if result != getMockSubject() {
		t.Error("Result does not match original serialized object.")
	}
}

func TestNewSerializerReturnsJSONSerializerForJSON5(t *testing.T) {
	content := getMockSubjectSerialize(t, JSONSerializer{})
	serializer, err := NewSerializer("example.json5", content)
	checkTestError(t, err)

	if reflect.TypeOf(serializer).Name() != "JSONSerializer" {
		t.Error("Got Serializer of wrong type when requesting JSONSerializer for .json5 file.")
	}
}

// Test INI serializer with various numeric types
type MockININumericTypes struct {
	UintVal   uint   `ini:"uint_val"`
	Uint8Val  uint8  `ini:"uint8_val"`
	FloatVal  float32 `ini:"float_val"`
	Float64Val float64 `ini:"float64_val"`
	BoolVal   bool   `ini:"bool_val"`
}

func TestINISerializerWithNumericTypes(t *testing.T) {
	serializer := INISerializer{}
	subject := MockININumericTypes{
		UintVal:   42,
		Uint8Val:  8,
		FloatVal:  3.14,
		Float64Val: 2.71828,
		BoolVal:   true,
	}

	serialized, err := serializer.Serialize(subject)
	checkTestError(t, err)

	if len(serialized) == 0 {
		t.Error("Expected non-empty serialized output")
	}
}

func TestINISerializerDeserializeInvalidINI(t *testing.T) {
	serializer := INISerializer{}

	// Invalid INI content (not valid key=value format)
	invalidINI := []byte("[section\nkey")

	result := MockSubject{}
	err := serializer.Deserialize(invalidINI, &result)
	if err == nil {
		t.Error("Expected error for invalid INI content")
	}
}
