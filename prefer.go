package prefer

import "log"

type filterable func(identifier string) bool

type Configuration struct {
	Identifier string

	Loaders     map[Loader]filterable
	Serializers map[Serializer]SerializerFactory
}

func Load(identifier string, out interface{}) (*Configuration, error) {
	configuration := NewConfiguration(identifier)
	return configuration, configuration.Reload(out)
}

func NewConfiguration(identifier string) *Configuration {
	return &Configuration{
		Identifier: identifier,
	}
}

func (configuration *Configuration) Reload(out interface{}) error {
	loader, err := NewLoader(configuration.Identifier)
	if err != nil {
		return err
	}

	identifier, content, err := loader.Load(configuration.Identifier)
	if err != nil {
		return err
	}

	configuration.Identifier = identifier
	log.Println(configuration.Identifier)

	serializer, err := NewSerializer(identifier, content)
	if err != nil {
		return err
	}

	return serializer.Deserialize(content, out)
}
