package prefer

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"path"

	"gopkg.in/h2non/filetype.v0"
)

// NOTE: It may make more sense to use a map to these instead of creating
// potentially unnecessray structs for implementing interfaces on.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type JSONSerializer struct{}
type XMLSerializer struct{}

type SerializerFactory func() Serializer

var defaultSerializers map[string]SerializerFactory

func NewSerializer(identifier string, content []byte) (serializer Serializer, err error) {
	var extension string

	if kind, unknown := filetype.Match(content); err == nil && unknown == nil && kind.Extension != "unknown" {
		extension = kind.Extension
	} else {
		extension = path.Ext(identifier)
	}

	factory, ok := defaultSerializers[extension]

	if !ok {
		return nil, errors.New("No matching serializer for " + identifier)
	}

	return factory(), nil
}

func NewJSONSerializer() Serializer {
	return JSONSerializer{}
}

func NewXMLSerializer() Serializer {
	return XMLSerializer{}
}

func (this JSONSerializer) Serialize(input interface{}) ([]byte, error) {
	return json.Marshal(input)
}

func (this JSONSerializer) Deserialize(input []byte, obj interface{}) error {
	return json.Unmarshal(input, &obj)
}

func (this XMLSerializer) Serialize(input interface{}) ([]byte, error) {
	return xml.Marshal(input)
}

func (this XMLSerializer) Deserialize(input []byte, obj interface{}) error {
	return xml.Unmarshal(input, &obj)
}

func init() {
	defaultSerializers = make(map[string]SerializerFactory)

	defaultSerializers[".json"] = NewJSONSerializer
	defaultSerializers[".xml"] = NewXMLSerializer
}
