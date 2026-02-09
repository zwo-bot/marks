package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFirefoxProfiles(t *testing.T) {
	// Create a mock Firefox profile structure
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, ".mozilla", "firefox", "abcdefgh.default-release")
	err := os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create places.sqlite to mark it as a valid profile
	err = os.WriteFile(filepath.Join(profileDir, "places.sqlite"), []byte{}, 0644)
	require.NoError(t, err)

	profiles := scanFirefoxProfiles(tmpDir)
	assert.NotEmpty(t, profiles)
	assert.Equal(t, "Firefox", profiles[0].Browser)
	assert.True(t, profiles[0].Selected)
}

func TestScanFirefoxProfilesWithProfilesIni(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base directory
	baseDir := filepath.Join(tmpDir, ".mozilla", "firefox")
	err := os.MkdirAll(baseDir, 0755)
	require.NoError(t, err)

	// Create profile directory
	profileDir := filepath.Join(baseDir, "test.profile")
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(profileDir, "places.sqlite"), []byte{}, 0644)
	require.NoError(t, err)

	// Write profiles.ini
	profilesIni := `[Profile0]
Name=test-profile
IsRelative=1
Path=test.profile
Default=1
`
	err = os.WriteFile(filepath.Join(baseDir, "profiles.ini"), []byte(profilesIni), 0644)
	require.NoError(t, err)

	profiles := scanFirefoxProfiles(tmpDir)
	assert.NotEmpty(t, profiles)
	assert.Equal(t, "test-profile", profiles[0].Name)
}

func TestScanFirefoxProfilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	profiles := scanFirefoxProfiles(tmpDir)
	assert.Empty(t, profiles)
}

func TestScanChromeProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Chrome Default profile
	defaultDir := filepath.Join(tmpDir, ".config", "google-chrome", "Default")
	err := os.MkdirAll(defaultDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(defaultDir, "Bookmarks"), []byte(`{}`), 0644)
	require.NoError(t, err)

	profiles := scanChromeProfiles(tmpDir)
	assert.NotEmpty(t, profiles)
	assert.Equal(t, "Chrome", profiles[0].Browser)
	assert.Equal(t, "Default", profiles[0].Name)
	assert.True(t, profiles[0].Selected)
}

func TestScanChromeProfilesMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Chrome Default + Profile 1
	chromeDir := filepath.Join(tmpDir, ".config", "google-chrome")
	for _, name := range []string{"Default", "Profile 1"} {
		dir := filepath.Join(chromeDir, name)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "Bookmarks"), []byte(`{}`), 0644)
		require.NoError(t, err)
	}

	profiles := scanChromeProfiles(tmpDir)
	assert.Len(t, profiles, 2)
}

func TestScanChromeProfilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	profiles := scanChromeProfiles(tmpDir)
	assert.Empty(t, profiles)
}

func TestParseFirefoxProfilesIni(t *testing.T) {
	tmpDir := t.TempDir()

	iniContent := `[General]
StartWithLastProfile=1

[Profile0]
Name=default-release
IsRelative=1
Path=Profiles/abc.default-release

[Profile1]
Name=dev-edition
IsRelative=0
Path=/opt/firefox/profiles/dev
`
	iniPath := filepath.Join(tmpDir, "profiles.ini")
	err := os.WriteFile(iniPath, []byte(iniContent), 0644)
	require.NoError(t, err)

	results := parseFirefoxProfilesIni(tmpDir, iniPath)
	assert.Len(t, results, 2)
	assert.Equal(t, "default-release", results[0].name)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/abc.default-release"), results[0].path)
	assert.Equal(t, "dev-edition", results[1].name)
	assert.Equal(t, "/opt/firefox/profiles/dev", results[1].path)
}

func TestTruncatePath(t *testing.T) {
	short := "/home/user/test"
	assert.Equal(t, short, truncatePath(short, 50))

	long := "/home/user/very/long/path/to/some/deep/nested/directory/file"
	result := truncatePath(long, 30)
	assert.LessOrEqual(t, len(result), 35) // allow for multi-byte '…'
	assert.Contains(t, result, "…")
}

func TestConfigFilePath(t *testing.T) {
	path, err := configFilePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.True(t, filepath.IsAbs(path))
}
