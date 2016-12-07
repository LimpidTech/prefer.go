package prefer

type filterable func(identifier string) bool

type Configuration struct {
	Identifier string

	Loaders     map[Loader]filterable
	Serializers map[Serializer]SerializerFactory
}

func Load(identifier string, out interface{}) (*Configuration, error) {
	this := NewConfiguration(identifier)
	return this, this.Reload(out)
}

func Watch(identifier string, out interface{}) (chan interface{}, error) {
	channel := make(chan interface{})
	configuration := NewConfiguration(identifier)
	go configuration.Watch(out, channel)
	return channel, nil
}

func NewConfiguration(identifier string) *Configuration {
	return &Configuration{
		Identifier: identifier,
	}
}

func (this *Configuration) Reload(out interface{}) error {
	loader, err := NewLoader(this.Identifier)
	if err != nil {
		return err
	}

	identifier, content, err := loader.Load()
	if err != nil {
		return err
	}

	this.Identifier = identifier

	serializer, err := NewSerializer(identifier, content)
	if err != nil {
		return err
	}

	return serializer.Deserialize(content, out)
}

func (this *Configuration) Watch(dest interface{}, channel chan interface{}) error {
	this.Reload(dest)
	channel <- dest

	update := make(chan bool)
	loader, err := NewLoader(this.Identifier)

	if err != nil {
		return err
	}

	go loader.Watch(update)

	for {
		if <-update == true {
			identifier, content, err := loader.Load()
			if err != nil {
				return err
			}

			serializer, err := NewSerializer(identifier, content)
			if err != nil {
				return err
			}

			err = serializer.Deserialize(content, dest)
			if err != nil {
				return err
			}

			channel <- dest
		}
	}

	return err
}
