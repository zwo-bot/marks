package config

import (
    "encoding/json"
    "fmt"
    "os"
    "os/user"
    "path/filepath"
    "runtime"
)


const appName = "bookmarks"


// GetConfigPath returns the path to the configuration file
func GetConfigPath(filename string) (string, error) {
    usr, err := user.Current()
    if err != nil {
        return "", err
    }

    var path string
    switch runtime.GOOS {
    case "linux":
        path = filepath.Join(usr.HomeDir, ".config", appName, filename) // Using .config directory
    case "darwin": // macOS is "darwin"
        path = filepath.Join(usr.HomeDir, "Library", "Application Support", "MyApp", filename)
    default:
        return "", fmt.Errorf("unsupported platform %s", runtime.GOOS)
    }

    return path, nil
}

// LoadConfig loads configuration from a file into the provided config struct
func LoadConfig(config interface{}, filename string) error {
    path, err := GetConfigPath(filename)
    if err != nil {
        return err
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, config)
}

// SaveConfig saves the provided config struct to a file
func SaveConfig(config interface{}, filename string) error {
    path, err := GetConfigPath(filename)
    if err != nil {
        return err
    }

    data, err := json.MarshalIndent(config, "", "    ") // Pretty print JSON
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644) // File permissions ensuring that the file is readable and writable by the user
}