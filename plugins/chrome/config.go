package chrome

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zwo-bot/marks/internal/logger"
)

type ChromeConfig struct {
	ProfilePath string `json:"profile_path"`
}

func (c *ChromeConfig) Load() error {
	log := logger.GetLogger()

	// If profile path is already set in config, verify it exists and is readable
	if c.ProfilePath != "" {
		log.Debug("Checking configured Chrome profile path", "path", c.ProfilePath)
		if _, err := os.Stat(c.ProfilePath); err == nil {
			// Try to open the file to verify we have read access
			if file, err := os.Open(c.ProfilePath); err == nil {
				file.Close()
				log.Debug("Using configured Chrome profile path", "path", c.ProfilePath)
				return nil
			}
		}
		log.Error("Configured Chrome profile path is not accessible", "path", c.ProfilePath)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %v", err)
	}

	// Try different possible Chrome paths
	possiblePaths := []string{
		// Standard Linux paths
		filepath.Join(home, ".config/google-chrome/Default/Bookmarks"),
		filepath.Join(home, ".config/chromium/Default/Bookmarks"),
		// Snap paths
		filepath.Join(home, "snap/chromium/common/chromium/Default/Bookmarks"),
		filepath.Join(home, "snap/chromium/common/.config/chromium/Default/Bookmarks"),
		filepath.Join(home, ".var/app/org.chromium.Chromium/config/chromium/Default/Bookmarks"),
		filepath.Join(home, "snap/google-chrome/current/.config/google-chrome/Default/Bookmarks"),
		filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome/Default/Bookmarks"),
	}

	for _, path := range possiblePaths {
		log.Debug("Checking Chrome profile path", "path", path)
		if _, err := os.Stat(path); err == nil {
			// Try to open the file to verify we have read access
			if file, err := os.Open(path); err == nil {
				file.Close()
				c.ProfilePath = path
				log.Debug("Found accessible Chrome profile", "path", path)
				return nil
			}
		}
	}

	return fmt.Errorf("no accessible Chrome profile found")
}

func (c *ChromeConfig) Save() error {
	// No need to save as we use default location
	return nil
}
