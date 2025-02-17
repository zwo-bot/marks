package config

import (
	"encoding/json"
	"os"
	"path/filepath"
    "os/user"
    "runtime"
    "fmt"
)

type AppConfig struct {
	Plugins        map[string]interface{} `json:"plugins,omitempty"`
	DefaultBrowser string                 `json:"defaultBrowser"`
}

// Global configuration variable
var GlobalConfig AppConfig

const configFilename = "app_config.json"

func InitializeConfig() error {
    configPath, err := GetConfigPath(configFilename)
    if err != nil {
        return err
    }

    // Check if the configuration file already exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        // File does not exist, create a default configuration with initialized map
        GlobalConfig = AppConfig{
            Plugins:        make(map[string]interface{}),
            DefaultBrowser: "firefox",  // Set the default browser or any default values
        }
        // Save the default configuration
        if err := SaveAppConfig(); err != nil {
            return fmt.Errorf("failed to save default configuration: %v", err)
        }
    } else {
        // File exists, load the existing configuration
        if err := LoadAppConfig(); err != nil {
            return err
        }
        // Make sure the Plugins map is initialized
        if GlobalConfig.Plugins == nil {
            GlobalConfig.Plugins = make(map[string]interface{})
        }
    }
    return nil
}



// GetConfigPath returns the path to the configuration file
func GetConfigPath(filename string) (string, error) {
    usr, err := user.Current()
    if err != nil {
        return "", err
    }

    var path string
    switch runtime.GOOS {
    case "linux":
        path = filepath.Join(usr.HomeDir, ".config", "bookmarks", filename) // Using .config directory
    case "darwin": // macOS is "darwin"
        path = filepath.Join(usr.HomeDir, "Library", "Application Support", "bookmarks", filename)
    default:
        return "", fmt.Errorf("unsupported platform %s", runtime.GOOS)
    }

    return path, nil
}

func LoadAppConfig() error {
	configPath, err := GetConfigPath(configFilename)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &GlobalConfig)
}

func SaveAppConfig() error {
	configPath, err := GetConfigPath(configFilename)
	if err != nil {
		return err
	}

	if err := ensureDir(configPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(GlobalConfig, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func ensureDir(filePath string) error {
    dir := filepath.Dir(filePath)
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        return os.MkdirAll(dir, 0755)
    }
    return nil
}