package prefer

import (
	"os"
	"strings"
)

// Source represents a configuration source for the ConfigBuilder.
type Source interface {
	// Load returns configuration data as a map.
	Load() (map[string]interface{}, error)
}

// DeepMerge merges override into base, returning a new map.
// Nested maps are merged recursively; other values are overwritten.
func DeepMerge(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Merge override
	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			baseMap, baseIsMap := baseVal.(map[string]interface{})
			overrideMap, overrideIsMap := v.(map[string]interface{})
			if baseIsMap && overrideIsMap {
				result[k] = DeepMerge(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

// ConfigBuilder builds configuration from multiple layered sources.
// Sources are applied in order, with later sources overriding earlier ones.
type ConfigBuilder struct {
	sources []Source
}

// NewConfigBuilder creates a new ConfigBuilder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		sources: make([]Source, 0),
	}
}

// AddSource adds a custom source to the builder.
func (b *ConfigBuilder) AddSource(source Source) *ConfigBuilder {
	b.sources = append(b.sources, source)
	return b
}

// AddDefaults adds in-memory default values.
func (b *ConfigBuilder) AddDefaults(defaults map[string]interface{}) *ConfigBuilder {
	return b.AddSource(&MemorySource{data: defaults})
}

// AddFile adds a required configuration file.
func (b *ConfigBuilder) AddFile(identifier string) *ConfigBuilder {
	return b.AddSource(&FileSource{identifier: identifier, required: true})
}

// AddOptionalFile adds an optional configuration file.
// If the file doesn't exist, it's silently skipped.
func (b *ConfigBuilder) AddOptionalFile(identifier string) *ConfigBuilder {
	return b.AddSource(&FileSource{identifier: identifier, required: false})
}

// AddEnv adds environment variables with the given prefix.
// Variables are converted to nested structure using the separator.
// Example: MYAPP__DATABASE__HOST with prefix "MYAPP" becomes database.host
func (b *ConfigBuilder) AddEnv(prefix string) *ConfigBuilder {
	return b.AddSource(&EnvSource{prefix: prefix, separator: "__"})
}

// AddEnvWithSeparator adds environment variables with a custom separator.
func (b *ConfigBuilder) AddEnvWithSeparator(prefix, separator string) *ConfigBuilder {
	return b.AddSource(&EnvSource{prefix: prefix, separator: separator})
}

// Build loads and merges all sources, returning a ConfigMap.
func (b *ConfigBuilder) Build() (*ConfigMap, error) {
	merged := make(map[string]interface{})

	for _, source := range b.sources {
		data, err := source.Load()
		if err != nil {
			return nil, err
		}
		merged = DeepMerge(merged, data)
	}

	return &ConfigMap{data: merged}, nil
}

// MemorySource provides configuration from an in-memory map.
type MemorySource struct {
	data map[string]interface{}
}

// NewMemorySource creates a new MemorySource.
func NewMemorySource(data map[string]interface{}) *MemorySource {
	return &MemorySource{data: data}
}

func (s *MemorySource) Load() (map[string]interface{}, error) {
	// Return a copy to prevent mutation
	result := make(map[string]interface{})
	for k, v := range s.data {
		result[k] = v
	}
	return result, nil
}

// FileSource loads configuration from a file.
type FileSource struct {
	identifier string
	required   bool
}

// NewFileSource creates a required file source.
func NewFileSource(identifier string) *FileSource {
	return &FileSource{identifier: identifier, required: true}
}

// NewOptionalFileSource creates an optional file source.
func NewOptionalFileSource(identifier string) *FileSource {
	return &FileSource{identifier: identifier, required: false}
}

func (s *FileSource) Load() (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := Load(s.identifier, &result)
	if err != nil {
		if !s.required {
			// Return empty map for optional files that don't exist
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	return result, nil
}

// EnvSource loads configuration from environment variables.
type EnvSource struct {
	prefix    string
	separator string
}

// NewEnvSource creates a new EnvSource with the default separator "__".
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{prefix: prefix, separator: "__"}
}

// NewEnvSourceWithSeparator creates an EnvSource with a custom separator.
func NewEnvSourceWithSeparator(prefix, separator string) *EnvSource {
	return &EnvSource{prefix: prefix, separator: separator}
}

func (s *EnvSource) Load() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	prefix := s.prefix + s.separator

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Remove prefix and convert to nested structure
		key = strings.TrimPrefix(key, prefix)
		key = strings.ToLower(key)
		keyParts := strings.Split(key, s.separator)

		setNested(result, keyParts, value)
	}

	return result, nil
}

// setNested sets a value in a nested map structure.
func setNested(data map[string]interface{}, parts []string, value interface{}) {
	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}
			if nested, ok := current[part].(map[string]interface{}); ok {
				current = nested
			} else {
				// Overwrite non-map value with a new map
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		}
	}
}
