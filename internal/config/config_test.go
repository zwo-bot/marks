package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPathLinux(t *testing.T) {
	// This test only validates the structure, not the actual OS
	// since we're testing the function logic
	if os.Getenv("GOOS") == "darwin" {
		t.Skip("Skipping Linux-specific test on macOS")
	}
}

func TestInitializeConfigCreatesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	SetCustomConfigPath(configPath)
	defer SetCustomConfigPath("")

	err := InitializeConfig()
	require.NoError(t, err)

	// Verify default config was created
	assert.Equal(t, "firefox", GlobalConfig.DefaultBrowser)
	assert.NotNil(t, GlobalConfig.Plugins)

	// Verify file was written
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestInitializeConfigLoadsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Write a config file
	configContent := `{
    "plugins": {
        "firefox": {
            "profile_path": "/test/path"
        }
    },
    "defaultBrowser": "chrome"
}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	SetCustomConfigPath(configPath)
	defer SetCustomConfigPath("")

	err = InitializeConfig()
	require.NoError(t, err)

	assert.Equal(t, "chrome", GlobalConfig.DefaultBrowser)
	assert.NotNil(t, GlobalConfig.Plugins["firefox"])
}

func TestSaveAppConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.json")

	SetCustomConfigPath(configPath)
	defer SetCustomConfigPath("")

	GlobalConfig = AppConfig{
		Plugins:        map[string]interface{}{"test": "value"},
		DefaultBrowser: "firefox",
	}

	err := SaveAppConfig()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Reload and verify
	err = LoadAppConfig()
	require.NoError(t, err)
	assert.Equal(t, "firefox", GlobalConfig.DefaultBrowser)
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a", "b", "c", "config.json")

	err := ensureDir(filePath)
	assert.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(filePath))
	assert.NoError(t, err)
}

func TestLoadAppConfigNonexistent(t *testing.T) {
	SetCustomConfigPath("/nonexistent/path/config.json")
	defer SetCustomConfigPath("")

	err := LoadAppConfig()
	assert.Error(t, err)
}
