package chrome

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zwo-bot/marks/internal/logger"
)

type ChromeConfig struct {
	ProfilePath string `json:"profile_path"`
}

func (c *ChromeConfig) Load() error {
	log := logger.GetLogger()

	if c.ProfilePath != "" {
		log.Debug("Checking configured Chrome profile path", "path", c.ProfilePath)
		if _, err := os.Stat(c.ProfilePath); err == nil {
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

	possiblePaths := chromeBookmarkPaths(home)

	for _, path := range possiblePaths {
		log.Debug("Checking Chrome profile path", "path", path)
		if _, err := os.Stat(path); err == nil {
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
	return nil
}

// chromeBookmarkPaths returns all possible Chrome/Chromium bookmark file paths.
func chromeBookmarkPaths(home string) []string {
	paths := linuxChromePaths(home)
	if runtime.GOOS == "darwin" {
		paths = macOSChromePaths(home)
	}
	return paths
}

// linuxChromePaths returns Chrome/Chromium bookmark file paths on Linux.
func linuxChromePaths(home string) []string {
	return []string{
		filepath.Join(home, ".config/google-chrome/Default/Bookmarks"),
		filepath.Join(home, ".config/chromium/Default/Bookmarks"),
		filepath.Join(home, "snap/chromium/common/chromium/Default/Bookmarks"),
		filepath.Join(home, "snap/chromium/common/.config/chromium/Default/Bookmarks"),
		filepath.Join(home, ".var/app/org.chromium.Chromium/config/chromium/Default/Bookmarks"),
		filepath.Join(home, "snap/google-chrome/current/.config/google-chrome/Default/Bookmarks"),
		filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome/Default/Bookmarks"),
	}
}

// macOSChromePaths returns Chrome/Chromium bookmark file paths on macOS.
func macOSChromePaths(home string) []string {
	return []string{
		filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "Bookmarks"),
		filepath.Join(home, "Library", "Application Support", "Chromium", "Default", "Bookmarks"),
	}
}

// ChromeBookmarkPaths returns all possible Chrome/Chromium bookmark file paths
// for the current OS. Exported for use by the configure command scanner.
func ChromeBookmarkPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return chromeBookmarkPaths(home)
}
