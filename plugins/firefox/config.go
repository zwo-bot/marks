package firefox

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/ini.v1"
)

type FirefoxConfig struct {
	ProfilePath string `json:"profile_path"`
}

func (c *FirefoxConfig) Load() error {
	// If profile path is already set (from config file), verify it exists
	if c.ProfilePath != "" {
		if _, err := os.Stat(c.ProfilePath); err == nil {
			return nil // Use the configured path
		}
	}

	// Otherwise, try to auto-detect
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	possiblePaths := linuxFirefoxPaths(home)
	if runtime.GOOS == "darwin" {
		possiblePaths = macOSFirefoxPaths(home)
	}

	for _, basePath := range possiblePaths {
		if _, err := os.Stat(basePath); err == nil {
			profilePath, err := getProfilePath(basePath)
			if err == nil {
				c.ProfilePath = profilePath
				return nil
			}
		}
	}

	// If no path is found, use a default but don't error
	if runtime.GOOS == "darwin" {
		c.ProfilePath = filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles", "default")
	} else {
		c.ProfilePath = filepath.Join(home, ".mozilla/firefox/default")
	}
	return nil
}

func (c *FirefoxConfig) Save() error {
	return nil
}

// linuxFirefoxPaths returns possible Firefox base directories on Linux.
func linuxFirefoxPaths(home string) []string {
	return []string{
		filepath.Join(home, ".mozilla/firefox"),
		filepath.Join(home, "snap/firefox/common/.mozilla/firefox"),
		filepath.Join(home, ".var/app/org.mozilla.firefox/.mozilla/firefox"),
	}
}

// macOSFirefoxPaths returns possible Firefox base directories on macOS.
func macOSFirefoxPaths(home string) []string {
	return []string{
		filepath.Join(home, "Library", "Application Support", "Firefox"),
	}
}

// FirefoxBasePaths returns all possible Firefox base directories for the current OS.
// Exported for use by the configure command scanner.
func FirefoxBasePaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	if runtime.GOOS == "darwin" {
		return macOSFirefoxPaths(home)
	}
	return linuxFirefoxPaths(home)
}

func getProfilePath(ffDir string) (string, error) {
	// Try installs.ini first
	profilePath, err := getProfileFromInstalls(ffDir)
	if err == nil {
		return profilePath, nil
	}

	// Fallback to profiles.ini
	return getProfileFromProfiles(ffDir)
}

func getProfileFromInstalls(ffDir string) (string, error) {
	installsPath := filepath.Join(ffDir, "installs.ini")
	cfg, err := ini.Load(installsPath)
	if err != nil {
		return "", err
	}

	for _, sec := range cfg.Sections() {
		if sec.HasKey("Default") {
			name, err := sec.GetKey("Default")
			if err != nil {
				return "", err
			}
			profilePath := name.String()
			if filepath.IsAbs(profilePath) {
				return profilePath, nil
			}
			return filepath.Join(ffDir, profilePath), nil
		}
	}

	return "", fmt.Errorf("no default profile found in installs.ini")
}

func getProfileFromProfiles(ffDir string) (string, error) {
	profilesPath := filepath.Join(ffDir, "profiles.ini")
	cfg, err := ini.Load(profilesPath)
	if err != nil {
		return "", err
	}

	// Look for profiles with Default=1
	for _, sec := range cfg.Sections() {
		if sec.HasKey("Default") && sec.HasKey("Path") {
			defaultKey, _ := sec.GetKey("Default")
			if defaultKey.String() == "1" {
				pathKey, _ := sec.GetKey("Path")
				profilePath := pathKey.String()
				isRelative := true
				if sec.HasKey("IsRelative") {
					relKey, _ := sec.GetKey("IsRelative")
					isRelative = relKey.String() == "1"
				}
				if isRelative {
					return filepath.Join(ffDir, profilePath), nil
				}
				return profilePath, nil
			}
		}
	}

	// Fallback: return the first profile found
	for _, sec := range cfg.Sections() {
		if sec.HasKey("Path") {
			pathKey, _ := sec.GetKey("Path")
			profilePath := pathKey.String()
			isRelative := true
			if sec.HasKey("IsRelative") {
				relKey, _ := sec.GetKey("IsRelative")
				isRelative = relKey.String() == "1"
			}
			if isRelative {
				return filepath.Join(ffDir, profilePath), nil
			}
			return profilePath, nil
		}
	}

	return "", fmt.Errorf("no profile found in profiles.ini")
}
