package prefer

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"reflect"

	"github.com/yosuke-furukawa/json5/encoding/json5"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// NOTE: It may make more sense to use a map to these instead of creating
// potentially unnecessray structs for implementing interfaces on.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type YAMLSerializer struct{}
type XMLSerializer struct{}
type INISerializer struct{}
type JSONSerializer struct{}

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

func NewINISerializer() Serializer {
	return INISerializer{}
}

func NewJSONSerializer() Serializer {
	return JSONSerializer{}
}

func (this YAMLSerializer) Serialize(input interface{}) ([]byte, error) {
	return yaml.Marshal(input)
}

func (this YAMLSerializer) Deserialize(input []byte, obj interface{}) error {
	return yaml.Unmarshal(input, obj)
}

func (this XMLSerializer) Serialize(input interface{}) ([]byte, error) {
	return xml.Marshal(input)
}

func (this XMLSerializer) Deserialize(input []byte, obj interface{}) error {
	return xml.Unmarshal(input, &obj)
}

func (this INISerializer) Serialize(input interface{}) ([]byte, error) {
	cfg := ini.Empty()
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)
	
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		
		var strValue string
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strValue = fmt.Sprintf("%d", value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strValue = fmt.Sprintf("%d", value.Uint())
		case reflect.Float32, reflect.Float64:
			strValue = fmt.Sprintf("%g", value.Float())
		case reflect.Bool:
			strValue = fmt.Sprintf("%t", value.Bool())
		default:
			strValue = value.String()
		}
		
		cfg.Section("").Key(field.Name).SetValue(strValue)
	}
	
	var buf bytes.Buffer
	_, err := cfg.WriteTo(&buf)
	return buf.Bytes(), err
}

func (this INISerializer) Deserialize(input []byte, obj interface{}) error {
	cfg, err := ini.Load(input)
	if err != nil {
		return err
	}
	return cfg.MapTo(obj)
}

func (this JSONSerializer) Serialize(input interface{}) ([]byte, error) {
	return json5.Marshal(input)
}

func (this JSONSerializer) Deserialize(input []byte, obj interface{}) error {
	return json5.Unmarshal(input, obj)
}

func init() {
	defaultSerializers = make(map[string]SerializerFactory)

	defaultSerializers[".json"] = NewJSONSerializer
	defaultSerializers[".yml"] = NewYAMLSerializer
	defaultSerializers[".yaml"] = NewYAMLSerializer
	defaultSerializers[".xml"] = NewXMLSerializer
	defaultSerializers[".ini"] = NewINISerializer
}
