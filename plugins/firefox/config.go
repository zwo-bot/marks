package firefox

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/ini.v1"
)

type FirefoxConfig struct {
	ProfilePath string `json:"profile_path"`
}

func (c *FirefoxConfig) Load() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Try different possible Firefox paths
	possiblePaths := []string{
		filepath.Join(home, ".mozilla/firefox"),
		filepath.Join(home, "snap/firefox/common/.mozilla/firefox"),
		filepath.Join(home, ".var/app/org.mozilla.firefox/.mozilla/firefox"),
	}

	for _, basePath := range possiblePaths {
		// Check if the directory exists
		if _, err := os.Stat(basePath); err == nil {
			// Try to get the default profile
			profilePath, err := getProfilePath(basePath)
			if err == nil {
				c.ProfilePath = profilePath
				return nil
			}
		}
	}

	// If no path is found, use a default but don't error
	c.ProfilePath = filepath.Join(home, ".mozilla/firefox/default")
	return nil
}

func (c *FirefoxConfig) Save() error {
	return nil
}

func getProfilePath(ffDir string) (string, error) {
	// Try to load installs.ini
	cfg, err := ini.Load(ffDir + "/installs.ini")
	if err != nil {
		return "", err
	}

	// Look for default profile
	for _, sec := range cfg.Sections() {
		if sec.HasKey("Default") {
			name, err := sec.GetKey("Default")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s/%s", ffDir, name), nil
		}
	}

	return "", fmt.Errorf("no default profile found")
}
