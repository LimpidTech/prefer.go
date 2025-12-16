package prefer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigMapGet(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
	}
	cm := NewConfigMap(data)

	// Test top-level key
	val, ok := cm.Get("name")
	if !ok || val != "test" {
		t.Error("Expected to get 'test' for key 'name'")
	}

	// Test nested key
	val, ok = cm.Get("database.host")
	if !ok || val != "localhost" {
		t.Error("Expected to get 'localhost' for key 'database.host'")
	}

	// Test non-existent key
	_, ok = cm.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Test non-existent nested key
	_, ok = cm.Get("database.nonexistent")
	if ok {
		t.Error("Expected false for non-existent nested key")
	}

	// Test empty key returns entire map
	val, ok = cm.Get("")
	if !ok {
		t.Error("Expected true for empty key")
	}
	if _, isMap := val.(map[string]interface{}); !isMap {
		t.Error("Expected map for empty key")
	}
}

func TestConfigMapGetString(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test",
		"number": 42,
	}
	cm := NewConfigMap(data)

	str, ok := cm.GetString("name")
	if !ok || str != "test" {
		t.Error("Expected to get 'test' string")
	}

	_, ok = cm.GetString("number")
	if ok {
		t.Error("Expected false for non-string value")
	}

	_, ok = cm.GetString("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapGetInt(t *testing.T) {
	data := map[string]interface{}{
		"int_val":   42,
		"float_val": 3.14,
		"json_num":  float64(100), // JSON numbers are float64
		"string":    "not a number",
	}
	cm := NewConfigMap(data)

	val, ok := cm.GetInt("int_val")
	if !ok || val != 42 {
		t.Error("Expected to get 42 for int_val")
	}

	// JSON numbers come as float64
	val, ok = cm.GetInt("json_num")
	if !ok || val != 100 {
		t.Error("Expected to get 100 for json_num")
	}

	// Float gets truncated
	val, ok = cm.GetInt("float_val")
	if !ok || val != 3 {
		t.Error("Expected to get 3 for float_val")
	}

	_, ok = cm.GetInt("string")
	if ok {
		t.Error("Expected false for string value")
	}

	_, ok = cm.GetInt("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapGetFloat(t *testing.T) {
	data := map[string]interface{}{
		"float_val": 3.14,
		"int_val":   42,
		"string":    "not a number",
	}
	cm := NewConfigMap(data)

	val, ok := cm.GetFloat("float_val")
	if !ok || val != 3.14 {
		t.Error("Expected to get 3.14 for float_val")
	}

	val, ok = cm.GetFloat("int_val")
	if !ok || val != 42.0 {
		t.Error("Expected to get 42.0 for int_val")
	}

	_, ok = cm.GetFloat("string")
	if ok {
		t.Error("Expected false for string value")
	}
}

func TestConfigMapGetBool(t *testing.T) {
	data := map[string]interface{}{
		"enabled":  true,
		"disabled": false,
		"string":   "true",
	}
	cm := NewConfigMap(data)

	val, ok := cm.GetBool("enabled")
	if !ok || val != true {
		t.Error("Expected to get true for enabled")
	}

	val, ok = cm.GetBool("disabled")
	if !ok || val != false {
		t.Error("Expected to get false for disabled")
	}

	_, ok = cm.GetBool("string")
	if ok {
		t.Error("Expected false for string value")
	}
}

func TestConfigMapGetSlice(t *testing.T) {
	data := map[string]interface{}{
		"items":  []interface{}{"a", "b", "c"},
		"string": "not a slice",
	}
	cm := NewConfigMap(data)

	slice, ok := cm.GetSlice("items")
	if !ok || len(slice) != 3 {
		t.Error("Expected to get slice with 3 items")
	}

	_, ok = cm.GetSlice("string")
	if ok {
		t.Error("Expected false for string value")
	}
}

func TestConfigMapGetMap(t *testing.T) {
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"key": "value",
		},
		"string": "not a map",
	}
	cm := NewConfigMap(data)

	m, ok := cm.GetMap("nested")
	if !ok || m["key"] != "value" {
		t.Error("Expected to get nested map")
	}

	_, ok = cm.GetMap("string")
	if ok {
		t.Error("Expected false for string value")
	}
}

func TestConfigMapSet(t *testing.T) {
	cm := NewConfigMap(nil)

	// Set top-level value
	err := cm.Set("name", "test")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	val, _ := cm.GetString("name")
	if val != "test" {
		t.Error("Expected name to be 'test'")
	}

	// Set nested value (creates intermediate maps)
	err = cm.Set("database.host", "localhost")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	val, _ = cm.GetString("database.host")
	if val != "localhost" {
		t.Error("Expected database.host to be 'localhost'")
	}

	// Set deeply nested value
	err = cm.Set("a.b.c.d", "deep")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	val, _ = cm.GetString("a.b.c.d")
	if val != "deep" {
		t.Error("Expected a.b.c.d to be 'deep'")
	}

	// Empty key should error
	err = cm.Set("", "value")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Set on non-map should error
	cm.Set("scalar", "value")
	err = cm.Set("scalar.nested", "value")
	if err == nil {
		t.Error("Expected error when setting nested key on scalar")
	}
}

func TestConfigMapHas(t *testing.T) {
	data := map[string]interface{}{
		"exists": "yes",
		"nested": map[string]interface{}{
			"key": "value",
		},
	}
	cm := NewConfigMap(data)

	if !cm.Has("exists") {
		t.Error("Expected Has to return true for 'exists'")
	}

	if !cm.Has("nested.key") {
		t.Error("Expected Has to return true for 'nested.key'")
	}

	if cm.Has("nonexistent") {
		t.Error("Expected Has to return false for 'nonexistent'")
	}
}

func TestConfigMapKeys(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	cm := NewConfigMap(data)

	keys := cm.Keys()
	if len(keys) != 3 {
		t.Error("Expected 3 keys")
	}
}

func TestConfigMapData(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	cm := NewConfigMap(data)

	if cm.Data()["key"] != "value" {
		t.Error("Expected Data() to return underlying map")
	}
}

func TestLoadMap(t *testing.T) {
	// Create a temporary JSON file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	content := `{
		"name": "test-app",
		"database": {
			"host": "localhost",
			"port": 5432
		},
		"features": {
			"auth": {
				"enabled": true
			}
		}
	}`

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := LoadMap(tmpFile)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Test top-level access
	name, ok := cm.GetString("name")
	if !ok || name != "test-app" {
		t.Error("Expected name to be 'test-app'")
	}

	// Test nested access
	host, ok := cm.GetString("database.host")
	if !ok || host != "localhost" {
		t.Error("Expected database.host to be 'localhost'")
	}

	port, ok := cm.GetInt("database.port")
	if !ok || port != 5432 {
		t.Error("Expected database.port to be 5432")
	}

	// Test deeply nested access
	enabled, ok := cm.GetBool("features.auth.enabled")
	if !ok || enabled != true {
		t.Error("Expected features.auth.enabled to be true")
	}
}

func TestLoadMapError(t *testing.T) {
	_, err := LoadMap("nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConfigMapGetThroughNonMap(t *testing.T) {
	data := map[string]interface{}{
		"scalar": "value",
	}
	cm := NewConfigMap(data)

	// Trying to access nested key through scalar should return false
	_, ok := cm.Get("scalar.nested")
	if ok {
		t.Error("Expected false when accessing nested key through scalar")
	}
}

func TestNewConfigMapWithNil(t *testing.T) {
	cm := NewConfigMap(nil)
	if cm.data == nil {
		t.Error("Expected non-nil map even when initialized with nil")
	}

	// Should work normally
	err := cm.Set("key", "value")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
}

func TestConfigMapGetIntWithInt64(t *testing.T) {
	data := map[string]interface{}{
		"int64_val": int64(9223372036854775807),
	}
	cm := NewConfigMap(data)

	_, ok := cm.GetInt("int64_val")
	if !ok {
		t.Error("Expected to get int64 value as int")
	}
}

func TestConfigMapGetFloatWithInt64(t *testing.T) {
	data := map[string]interface{}{
		"int64_val": int64(100),
	}
	cm := NewConfigMap(data)

	val, ok := cm.GetFloat("int64_val")
	if !ok || val != 100.0 {
		t.Error("Expected to get int64 value as float64")
	}
}

func TestConfigMapGetFloatNonexistent(t *testing.T) {
	cm := NewConfigMap(nil)

	_, ok := cm.GetFloat("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapGetBoolNonexistent(t *testing.T) {
	cm := NewConfigMap(nil)

	_, ok := cm.GetBool("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapGetSliceNonexistent(t *testing.T) {
	cm := NewConfigMap(nil)

	_, ok := cm.GetSlice("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapGetMapNonexistent(t *testing.T) {
	cm := NewConfigMap(nil)

	_, ok := cm.GetMap("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
}

func TestConfigMapSetThroughExistingMap(t *testing.T) {
	// Test setting a nested value when the intermediate map already exists
	data := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
		},
	}
	cm := NewConfigMap(data)

	// Set a new key in an existing nested map
	err := cm.Set("database.port", 5432)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	port, ok := cm.GetInt("database.port")
	if !ok || port != 5432 {
		t.Error("Expected database.port to be 5432")
	}

	// Verify the existing value is still there
	host, ok := cm.GetString("database.host")
	if !ok || host != "localhost" {
		t.Error("Expected database.host to still be 'localhost'")
	}
}
