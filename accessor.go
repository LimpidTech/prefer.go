package prefer

import (
	"fmt"
	"strings"
)

const keySeparator = "."

// ConfigMap provides dot-notation access to configuration values.
// It wraps a map[string]interface{} and supports nested key access
// using dot-separated paths like "database.host".
type ConfigMap struct {
	data map[string]interface{}
}

// NewConfigMap creates a new ConfigMap from a map.
func NewConfigMap(data map[string]interface{}) *ConfigMap {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &ConfigMap{data: data}
}

// LoadMap loads a configuration file into a ConfigMap for dot-notation access.
func LoadMap(identifier string) (*ConfigMap, error) {
	data := make(map[string]interface{})
	_, err := Load(identifier, &data)
	if err != nil {
		return nil, err
	}
	return NewConfigMap(data), nil
}

// Get retrieves a value by dot-separated key path.
// Returns the value and true if found, nil and false otherwise.
func (c *ConfigMap) Get(key string) (interface{}, bool) {
	if key == "" {
		return c.data, true
	}

	parts := strings.Split(key, keySeparator)
	var current interface{} = c.data

	for _, part := range parts {
		switch node := current.(type) {
		case map[string]interface{}:
			val, ok := node[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// GetString retrieves a string value by key.
// Returns the value and true if found and is a string, empty string and false otherwise.
func (c *ConfigMap) GetString(key string) (string, bool) {
	val, ok := c.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an integer value by key.
// Handles both int and float64 (JSON numbers are float64).
func (c *ConfigMap) GetInt(key string) (int, bool) {
	val, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloat retrieves a float64 value by key.
func (c *ConfigMap) GetFloat(key string) (float64, bool) {
	val, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a boolean value by key.
func (c *ConfigMap) GetBool(key string) (bool, bool) {
	val, ok := c.Get(key)
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetSlice retrieves a slice value by key.
func (c *ConfigMap) GetSlice(key string) ([]interface{}, bool) {
	val, ok := c.Get(key)
	if !ok {
		return nil, false
	}
	slice, ok := val.([]interface{})
	return slice, ok
}

// GetMap retrieves a nested map value by key.
func (c *ConfigMap) GetMap(key string) (map[string]interface{}, bool) {
	val, ok := c.Get(key)
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	return m, ok
}

// Set sets a value at the given dot-separated key path.
// Creates intermediate maps as needed.
func (c *ConfigMap) Set(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	parts := strings.Split(key, keySeparator)
	current := c.data

	// Navigate to the parent of the final key, creating maps as needed
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			// Create intermediate map
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		} else if nextMap, ok := next.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return fmt.Errorf("cannot set %s: %s is not a map", key, strings.Join(parts[:i+1], keySeparator))
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
	return nil
}

// Has checks if a key exists.
func (c *ConfigMap) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Data returns the underlying map.
func (c *ConfigMap) Data() map[string]interface{} {
	return c.data
}

// Keys returns all top-level keys.
func (c *ConfigMap) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}
