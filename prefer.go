package prefer

type filterable func(identifier string) bool

// Option configures how configuration is loaded
type Option func(*Configuration)

// WithLoader sets a custom loader for the configuration
func WithLoader(loader Loader) Option {
	return func(c *Configuration) {
		c.loader = loader
	}
}

type Configuration struct {
	Identifier string

	loader      Loader
	Loaders     map[Loader]filterable
	Serializers map[Serializer]SerializerFactory
}

// Load loads configuration from the given identifier into dest.
// Options can be used to customize loading behavior, e.g., WithLoader.
func Load(identifier string, dest interface{}, opts ...Option) (*Configuration, error) {
	this := NewConfiguration(identifier, opts...)
	return this, this.Reload(dest)
}

// Watch watches for configuration changes and returns a channel that receives
// updated configuration values.
func Watch(identifier string, dest interface{}, opts ...Option) (chan interface{}, error) {
	return WatchWithDone(identifier, dest, nil, opts...)
}

// WatchWithDone watches a configuration file with support for graceful shutdown.
// Close the done channel to stop watching.
func WatchWithDone(identifier string, dest interface{}, done <-chan struct{}, opts ...Option) (chan interface{}, error) {
	channel := make(chan interface{})
	configuration := NewConfiguration(identifier, opts...)
	go func() {
		_ = configuration.WatchWithDone(dest, channel, done)
	}()
	return channel, nil
}

// NewConfiguration creates a new Configuration with the given identifier and options.
func NewConfiguration(identifier string, opts ...Option) *Configuration {
	c := &Configuration{
		Identifier: identifier,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (this *Configuration) Reload(dest interface{}) error {
	loader := this.loader
	if loader == nil {
		var err error
		loader, err = NewLoader(this.Identifier)
		if err != nil {
			return err
		}
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
	loader := this.loader
	if loader == nil {
		var err error
		loader, err = NewLoader(this.Identifier)
		if err != nil {
			return err
		}
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
