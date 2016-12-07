package prefer

type filterable func(identifier string) bool

type Configuration struct {
	identifier  string
	loaders     map[Loader]filterable
	serializers map[Serializer]filterable
}

func Load(identifier string, out interface{}) (*Configuration, error) {
	configuration := NewConfiguration(identifier)
	err := configuration.Reload(out)
	return configuration, err
}

func NewConfiguration(identifier string) *Configuration {
	return &Configuration{
		identifier: identifier,
	}
}

func (configuration *Configuration) Reload(out interface{}) error {
	loader, err := NewLoader(configuration.identifier)
	if err != nil {
		return err
	}

	content, err := loader.Load(configuration.identifier)
	if err != nil {
		return err
	}

	serializer, err := NewSerializer(configuration.identifier, content)
	if err != nil {
		return err
	}

	err = serializer.Deserialize(content, out)
	return err
}
