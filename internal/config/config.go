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

const configFilename = "config.json"

var customConfigPath string

func SetCustomConfigPath(path string) {
    customConfigPath = path
}

func InitializeConfig() error {
    var configPath string

    if customConfigPath != "" {
        configPath = customConfigPath
    } else {
        // Try current directory first
        if _, err := os.Stat(configFilename); err == nil {
            var err error
            configPath, err = filepath.Abs(configFilename)
            if err != nil {
                return err
            }
        } else {
            // If not in current directory, try XDG paths
            var err error
            configPath, err = GetConfigPath(configFilename)
            if err != nil {
                return err
            }
        }
    }

    // Check if the configuration file already exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        // File does not exist, create a default configuration with initialized map
        GlobalConfig = AppConfig{
            Plugins:        make(map[string]interface{}),
            DefaultBrowser: "firefox",
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
    // First check current directory
    if _, err := os.Stat(filename); err == nil {
        absPath, err := filepath.Abs(filename)
        if err != nil {
            return "", err
        }
        return absPath, nil
    }

    // If not found in current directory, check XDG paths
    usr, err := user.Current()
    if err != nil {
        return "", err
    }

    var path string
    switch runtime.GOOS {
    case "linux":
        path = filepath.Join(usr.HomeDir, ".config", "marks", filename)
    case "darwin": // macOS is "darwin"
        path = filepath.Join(usr.HomeDir, "Library", "Application Support", "marks", filename)
    default:
        return "", fmt.Errorf("unsupported platform %s", runtime.GOOS)
    }

    return path, nil
}

func LoadAppConfig() error {
    var configPath string
    var err error

    if customConfigPath != "" {
        configPath = customConfigPath
    } else {
        // Try current directory first
        if _, err := os.Stat(configFilename); err == nil {
            configPath, err = filepath.Abs(configFilename)
            if err != nil {
                return err
            }
        } else {
            // If not in current directory, try XDG paths
            configPath, err = GetConfigPath(configFilename)
            if err != nil {
                return err
            }
        }
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, &GlobalConfig)
}

func SaveAppConfig() error {
    var configPath string
    var err error

    if customConfigPath != "" {
        configPath = customConfigPath
    } else {
        // Try current directory first if config exists there
        if _, err := os.Stat(configFilename); err == nil {
            configPath, err = filepath.Abs(configFilename)
            if err != nil {
                return err
            }
        } else {
            // If not in current directory or doesn't exist, use XDG paths
            configPath, err = GetConfigPath(configFilename)
            if err != nil {
                return err
            }
        }
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
