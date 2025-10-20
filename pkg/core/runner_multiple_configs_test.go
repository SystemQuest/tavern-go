package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadGlobalConfigs tests loading and merging multiple global config files
// Aligned with tavern-py commit 76569fd: load_global_config()
func TestLoadGlobalConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first global config file
	config1Path := filepath.Join(tmpDir, "config1.yaml")
	config1 := `
variables:
  base_url: "http://api1.example.com"
  api_key: "key1"
  timeout: 30

settings:
  retry: 3
  log_level: "info"
`
	err := os.WriteFile(config1Path, []byte(config1), 0644)
	require.NoError(t, err)

	// Create second global config file (will override some values)
	config2Path := filepath.Join(tmpDir, "config2.yaml")
	config2 := `
variables:
  base_url: "http://api2.example.com"
  new_var: "new_value"

settings:
  retry: 5
  new_setting: "value"
`
	err = os.WriteFile(config2Path, []byte(config2), 0644)
	require.NoError(t, err)

	// Create third global config file
	config3Path := filepath.Join(tmpDir, "config3.yaml")
	config3 := `
variables:
  timeout: 60
  third_var: "third"
`
	err = os.WriteFile(config3Path, []byte(config3), 0644)
	require.NoError(t, err)

	// Create runner
	runner, err := NewRunner(&Config{
		BaseDir: tmpDir,
	})
	require.NoError(t, err)

	// Load multiple global configs
	err = runner.LoadGlobalConfigs([]string{config1Path, config2Path, config3Path})
	require.NoError(t, err)

	// Verify merged variables
	// base_url: config2 should override config1
	assert.Equal(t, "http://api2.example.com", runner.config.Variables["base_url"])

	// api_key: only in config1, should be preserved
	assert.Equal(t, "key1", runner.config.Variables["api_key"])

	// timeout: config3 should override config1
	assert.Equal(t, 60, runner.config.Variables["timeout"])

	// new_var: only in config2
	assert.Equal(t, "new_value", runner.config.Variables["new_var"])

	// third_var: only in config3
	assert.Equal(t, "third", runner.config.Variables["third_var"])

	// Verify deep merge of settings
	globalCfg := runner.config.GlobalConfig
	settings, ok := globalCfg["settings"].(map[string]interface{})
	require.True(t, ok, "settings should be a map")

	// retry: config2 should override config1
	assert.Equal(t, 5, settings["retry"])

	// log_level: only in config1, should be preserved
	assert.Equal(t, "info", settings["log_level"])

	// new_setting: only in config2
	assert.Equal(t, "value", settings["new_setting"])
}

// TestLoadGlobalConfigsEmpty tests loading with empty file list
func TestLoadGlobalConfigsEmpty(t *testing.T) {
	runner, err := NewRunner(&Config{
		BaseDir: ".",
	})
	require.NoError(t, err)

	// Load with empty list should not error
	err = runner.LoadGlobalConfigs([]string{})
	require.NoError(t, err)

	// GlobalConfig should be initialized but empty
	assert.NotNil(t, runner.config.GlobalConfig)
}

// TestLoadGlobalConfigsSingle tests loading a single config file
func TestLoadGlobalConfigsSingle(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "config.yaml")
	config := `
variables:
  test_var: "test_value"
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	runner, err := NewRunner(&Config{
		BaseDir: tmpDir,
	})
	require.NoError(t, err)

	err = runner.LoadGlobalConfigs([]string{configPath})
	require.NoError(t, err)

	assert.Equal(t, "test_value", runner.config.Variables["test_var"])
}

// TestLoadGlobalConfigsError tests error handling
func TestLoadGlobalConfigsError(t *testing.T) {
	runner, err := NewRunner(&Config{
		BaseDir: ".",
	})
	require.NoError(t, err)

	// Try to load non-existent file
	err = runner.LoadGlobalConfigs([]string{"/nonexistent/file.yaml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load global config")
}
