package prefer

type filterable func(identifier string) bool

type Configuration struct {
	Identifier string

	Loaders     map[Loader]filterable
	Serializers map[Serializer]SerializerFactory
}

func Load(identifier string, dest interface{}) (*Configuration, error) {
	this := NewConfiguration(identifier)
	return this, this.Reload(dest)
}

func Watch(identifier string, dest interface{}) (chan interface{}, error) {
	return WatchWithDone(identifier, dest, nil)
}

// WatchWithDone watches a configuration file with support for graceful shutdown.
// Close the done channel to stop watching.
func WatchWithDone(identifier string, dest interface{}, done <-chan struct{}) (chan interface{}, error) {
	channel := make(chan interface{})
	configuration := NewConfiguration(identifier)
	go configuration.WatchWithDone(dest, channel, done)
	return channel, nil
}

func NewConfiguration(identifier string) *Configuration {
	return &Configuration{
		Identifier: identifier,
	}
}

func (this *Configuration) Reload(dest interface{}) error {
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

	return serializer.Deserialize(content, dest)
}

func (this *Configuration) Watch(dest interface{}, channel chan interface{}) error {
	return this.WatchWithDone(dest, channel, nil)
}

// WatchWithDone watches for configuration changes with support for graceful shutdown.
// Close the done channel to stop watching. Errors during reload are skipped
// (resilient watching) rather than terminating the watch loop.
func (this *Configuration) WatchWithDone(dest interface{}, channel chan interface{}, done <-chan struct{}) error {
	if err := this.Reload(dest); err != nil {
		return err
	}
	channel <- dest

	update := make(chan bool)
	loader, err := NewLoader(this.Identifier)

	if err != nil {
		return err
	}

	if err := loader.WatchWithContext(update, done); err != nil {
		return err
	}

	go func() {
		defer close(channel)
		for {
			select {
			case _, ok := <-update:
				if !ok {
					return
				}
				// Reload configuration - skip errors rather than terminating (resilient)
				identifier, content, err := loader.Load()
				if err != nil {
					continue
				}

				serializer, err := NewSerializer(identifier, content)
				if err != nil {
					continue
				}

				if err = serializer.Deserialize(content, dest); err != nil {
					continue
				}

				this.Identifier = identifier
				channel <- dest
			case <-done:
				return
			}
		}
	}()

	return nil
}
