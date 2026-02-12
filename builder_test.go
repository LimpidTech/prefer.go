package prefer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeepMerge(t *testing.T) {
	base := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
		"debug": true,
	}

	override := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "production.example.com",
		},
		"timeout": 30,
	}

	result := DeepMerge(base, override)

	// Check nested merge
	db := result["database"].(map[string]interface{})
	if db["host"] != "production.example.com" {
		t.Error("Expected host to be overridden")
	}
	if db["port"] != 5432 {
		t.Error("Expected port to be preserved from base")
	}

	// Check top-level values
	if result["debug"] != true {
		t.Error("Expected debug to be preserved from base")
	}
	if result["timeout"] != 30 {
		t.Error("Expected timeout from override")
	}
}

func TestDeepMergeNonMapOverride(t *testing.T) {
	base := map[string]interface{}{
		"value": map[string]interface{}{"nested": true},
	}

	override := map[string]interface{}{
		"value": "string now",
	}

	result := DeepMerge(base, override)
	if result["value"] != "string now" {
		t.Error("Expected non-map to override map")
	}
}

func TestConfigBuilderWithDefaults(t *testing.T) {
	builder := NewConfigBuilder().
		AddDefaults(map[string]interface{}{
			"host": "localhost",
			"port": 8080,
		})

	config, err := builder.Build()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	host, ok := config.GetString("host")
	if !ok || host != "localhost" {
		t.Error("Expected host to be localhost")
	}

	port, ok := config.GetInt("port")
	if !ok || port != 8080 {
		t.Error("Expected port to be 8080")
	}
}

func TestConfigBuilderLayeredOverride(t *testing.T) {
	builder := NewConfigBuilder().
		AddDefaults(map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
		}).
		AddSource(&MemorySource{data: map[string]interface{}{
			"database": map[string]interface{}{
				"host": "production.example.com",
			},
		}})

	config, err := builder.Build()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	// Host should be overridden
	host, _ := config.GetString("database.host")
	if host != "production.example.com" {
		t.Error("Expected host to be overridden")
	}

	// Port should still be from defaults
	port, _ := config.GetInt("database.port")
	if port != 5432 {
		t.Error("Expected port from defaults")
	}
}

func TestConfigBuilderWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	content := `{"name": "from file", "version": 1}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := NewConfigBuilder().
		AddDefaults(map[string]interface{}{"name": "default"}).
		AddFile(tmpFile).
		Build()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	name, _ := config.GetString("name")
	if name != "from file" {
		t.Error("Expected name from file, got:", name)
	}
}

func TestConfigBuilderWithOptionalFileMissing(t *testing.T) {
	config, err := NewConfigBuilder().
		AddDefaults(map[string]interface{}{"name": "default"}).
		AddOptionalFile("/nonexistent/file.json").
		Build()

	if err != nil {
		t.Fatal("Optional file should not cause error:", err)
	}

	name, _ := config.GetString("name")
	if name != "default" {
		t.Error("Expected name from defaults")
	}
}

func TestConfigBuilderWithEnv(t *testing.T) {
	// Set test env vars
	os.Setenv("TESTAPP__DATABASE__HOST", "envhost")
	os.Setenv("TESTAPP__DATABASE__PORT", "9999")
	os.Setenv("TESTAPP__DEBUG", "true")
	defer func() {
		os.Unsetenv("TESTAPP__DATABASE__HOST")
		os.Unsetenv("TESTAPP__DATABASE__PORT")
		os.Unsetenv("TESTAPP__DEBUG")
	}()

	config, err := NewConfigBuilder().
		AddDefaults(map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
			},
		}).
		AddEnv("TESTAPP").
		Build()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	host, _ := config.GetString("database.host")
	if host != "envhost" {
		t.Error("Expected host from env, got:", host)
	}

	debug, _ := config.GetString("debug")
	if debug != "true" {
		t.Error("Expected debug from env")
	}
}

func TestEnvSourceLoad(t *testing.T) {
	os.Setenv("TEST__DB__HOST", "localhost")
	os.Setenv("TEST__DB__PORT", "5432")
	defer func() {
		os.Unsetenv("TEST__DB__HOST")
		os.Unsetenv("TEST__DB__PORT")
	}()

	source := NewEnvSource("TEST")
	data, err := source.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	db := data["db"].(map[string]interface{})
	if db["host"] != "localhost" {
		t.Error("Expected host to be localhost")
	}
	if db["port"] != "5432" {
		t.Error("Expected port to be 5432")
	}
}

func TestMemorySourceLoad(t *testing.T) {
	source := NewMemorySource(map[string]interface{}{
		"key": "value",
	})

	data, err := source.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if data["key"] != "value" {
		t.Error("Expected key to be value")
	}

	// Verify it returns a copy
	data["key"] = "modified"
	data2, _ := source.Load()
	if data2["key"] != "value" {
		t.Error("MemorySource should return a copy")
	}
}

func TestFileSourceLoad(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	content := `{"key": "value"}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	source := NewFileSource(tmpFile)
	data, err := source.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if data["key"] != "value" {
		t.Error("Expected key to be value")
	}
}

func TestOptionalFileSourceMissing(t *testing.T) {
	source := NewOptionalFileSource("/nonexistent/file.json")
	data, err := source.Load()
	if err != nil {
		t.Fatal("Optional file should not error:", err)
	}

	if len(data) != 0 {
		t.Error("Expected empty map for missing optional file")
	}
}

func TestSetNested(t *testing.T) {
	data := make(map[string]interface{})
	setNested(data, []string{"a", "b", "c"}, "value")

	a := data["a"].(map[string]interface{})
	b := a["b"].(map[string]interface{})
	if b["c"] != "value" {
		t.Error("Expected nested value to be set")
	}
}

func TestSetNestedOverwritesNonMap(t *testing.T) {
	data := map[string]interface{}{
		"a": "string",
	}
	setNested(data, []string{"a", "b"}, "value")

	a := data["a"].(map[string]interface{})
	if a["b"] != "value" {
		t.Error("Expected string to be overwritten with map")
	}
}

func TestAddEnvWithSeparator(t *testing.T) {
	os.Setenv("APP-DB-HOST", "localhost")
	defer os.Unsetenv("APP-DB-HOST")

	config, err := NewConfigBuilder().
		AddEnvWithSeparator("APP", "-").
		Build()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	host, ok := config.GetString("db.host")
	if !ok || host != "localhost" {
		t.Error("Expected db.host to be localhost")
	}
}

func TestNewEnvSourceWithSeparator(t *testing.T) {
	os.Setenv("TEST2-A-B", "value")
	defer os.Unsetenv("TEST2-A-B")

	source := NewEnvSourceWithSeparator("TEST2", "-")
	data, err := source.Load()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	a := data["a"].(map[string]interface{})
	if a["b"] != "value" {
		t.Error("Expected nested value")
	}
}
