package prefer

import (
	"encoding/xml"
	"errors"
	"path"

	"gopkg.in/yaml.v2"
)

// NOTE: It may make more sense to use a map to these instead of creating
// potentially unnecessray structs for implementing interfaces on.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type YAMLSerializer struct{}
type XMLSerializer struct{}

type SerializerFactory func() Serializer

var defaultSerializers map[string]SerializerFactory

func NewSerializer(identifier string, content []byte) (serializer Serializer, err error) {
	extension := path.Ext(identifier)
	factory, ok := defaultSerializers[extension]

	if !ok {
		return nil, errors.New("No matching serializer for " + identifier)
	}

	return factory(), nil
}

func NewYAMLSerializer() Serializer {
	return YAMLSerializer{}
}

func NewXMLSerializer() Serializer {
	return XMLSerializer{}
}

func (this YAMLSerializer) Serialize(input interface{}) ([]byte, error) {
	return yaml.Marshal(input)
}

func (this YAMLSerializer) Deserialize(input []byte, obj interface{}) error {
	return yaml.Unmarshal(input, &obj)
}

func (this XMLSerializer) Serialize(input interface{}) ([]byte, error) {
	return xml.Marshal(input)
}

func (this XMLSerializer) Deserialize(input []byte, obj interface{}) error {
	return xml.Unmarshal(input, &obj)
}

func init() {
	defaultSerializers = make(map[string]SerializerFactory)

	defaultSerializers[".json"] = NewYAMLSerializer
	defaultSerializers[".yml"] = NewYAMLSerializer
	defaultSerializers[".yaml"] = NewYAMLSerializer
	defaultSerializers[".xml"] = NewXMLSerializer
}
