package favicon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/zwo-bot/marks/db"
)

// CacheDir returns the path to the favicon cache directory
func CacheDir() (string, error) {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %v", err)
		}
		cacheHome = filepath.Join(home, ".cache")
	}

	cacheDir := filepath.Join(cacheHome, "marks-favicons")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %v", err)
	}

	return cacheDir, nil
}

// SaveAndCacheIcon stores the icon in both the database and filesystem cache
// Returns the path to the cached file for use with rofi
func SaveAndCacheIcon(iconData []byte, urlStr string) (string, error) {
	if len(iconData) == 0 {
		return "", fmt.Errorf("no icon data provided")
	}

	// Save to database first
	favicon, err := db.SaveFavicon(iconData, urlStr)
	if err != nil {
		return "", fmt.Errorf("could not save favicon to database: %v", err)
	}

	// Get cache directory for filesystem storage
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	// Create hash of icon data for filename
	hash := sha256.Sum256(iconData)
	filename := hex.EncodeToString(hash[:])
	iconPath := filepath.Join(cacheDir, filename)

	// Check if icon already exists in filesystem cache
	if _, err := os.Stat(iconPath); err == nil {
		return iconPath, nil
	}

	// Write icon to filesystem cache
	if err := os.WriteFile(iconPath, favicon.Data, 0644); err != nil {
		return "", fmt.Errorf("could not write icon to cache: %v", err)
	}

	return iconPath, nil
}

// GetIconPath returns the filesystem path for a favicon, fetching from database if needed
func GetIconPath(urlStr string) (string, error) {
	// Parse URL to get domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	domain := parsedURL.Host

	// Try to get favicon from database
	favicon, err := db.GetFaviconByDomain(domain)
	if err != nil {
		return "", err
	}
	if favicon == nil {
		return "", nil
	}

	// Cache the icon data to filesystem if it exists in database
	return SaveAndCacheIcon(favicon.Data, urlStr)
}
