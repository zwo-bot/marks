package chrome

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxChromePaths(t *testing.T) {
	home := "/home/testuser"
	paths := linuxChromePaths(home)

	assert.Contains(t, paths, filepath.Join(home, ".config/google-chrome/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, ".config/chromium/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, "snap/chromium/common/chromium/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, "snap/chromium/common/.config/chromium/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, ".var/app/org.chromium.Chromium/config/chromium/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, "snap/google-chrome/current/.config/google-chrome/Default/Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, ".var/app/com.google.Chrome/config/google-chrome/Default/Bookmarks"))
}

func TestMacOSChromePaths(t *testing.T) {
	home := "/Users/testuser"
	paths := macOSChromePaths(home)

	assert.Contains(t, paths, filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "Bookmarks"))
	assert.Contains(t, paths, filepath.Join(home, "Library", "Application Support", "Chromium", "Default", "Bookmarks"))
	assert.Len(t, paths, 2)
}

func TestChromeConfigLoadWithValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	bookmarksPath := filepath.Join(tmpDir, "Bookmarks")
	err := os.WriteFile(bookmarksPath, []byte(`{"checksum":"","roots":{}}`), 0644)
	require.NoError(t, err)

	cfg := &ChromeConfig{ProfilePath: bookmarksPath}
	err = cfg.Load()
	assert.NoError(t, err)
	assert.Equal(t, bookmarksPath, cfg.ProfilePath)
}

func TestChromeConfigLoadWithInvalidPath(t *testing.T) {
	cfg := &ChromeConfig{ProfilePath: "/nonexistent/path/Bookmarks"}
	err := cfg.Load()
	assert.Error(t, err)
}

func TestChromeConfigSave(t *testing.T) {
	cfg := &ChromeConfig{}
	err := cfg.Save()
	assert.NoError(t, err)
}
