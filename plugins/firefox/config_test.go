package firefox

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfileFromInstalls(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()

	// Write a mock installs.ini
	installsContent := `[ABC123]
Default=Profiles/test-profile.default
Locked=1
`
	err := os.WriteFile(filepath.Join(tmpDir, "installs.ini"), []byte(installsContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfileFromInstalls(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/test-profile.default"), profilePath)
}

func TestGetProfileFromInstallsAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	installsContent := `[ABC123]
Default=/absolute/path/to/profile
Locked=1
`
	err := os.WriteFile(filepath.Join(tmpDir, "installs.ini"), []byte(installsContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfileFromInstalls(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "/absolute/path/to/profile", profilePath)
}

func TestGetProfileFromInstallsNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	installsContent := `[General]
Version=2
`
	err := os.WriteFile(filepath.Join(tmpDir, "installs.ini"), []byte(installsContent), 0644)
	require.NoError(t, err)

	_, err = getProfileFromInstalls(tmpDir)
	assert.Error(t, err)
}

func TestGetProfileFromProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	profilesContent := `[General]
StartWithLastProfile=1

[Profile0]
Name=default-release
IsRelative=1
Path=Profiles/abcdefgh.default-release
Default=1

[Profile1]
Name=other
IsRelative=1
Path=Profiles/12345678.other
`
	err := os.WriteFile(filepath.Join(tmpDir, "profiles.ini"), []byte(profilesContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfileFromProfiles(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/abcdefgh.default-release"), profilePath)
}

func TestGetProfileFromProfilesAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	profilesContent := `[Profile0]
Name=default
IsRelative=0
Path=/custom/path/to/profile
Default=1
`
	err := os.WriteFile(filepath.Join(tmpDir, "profiles.ini"), []byte(profilesContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfileFromProfiles(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "/custom/path/to/profile", profilePath)
}

func TestGetProfileFromProfilesFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// No Default=1, should return first profile found
	profilesContent := `[Profile0]
Name=some-profile
IsRelative=1
Path=Profiles/abcdefgh.some-profile
`
	err := os.WriteFile(filepath.Join(tmpDir, "profiles.ini"), []byte(profilesContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfileFromProfiles(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/abcdefgh.some-profile"), profilePath)
}

func TestGetProfileFromProfilesNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := getProfileFromProfiles(tmpDir)
	assert.Error(t, err)
}

func TestGetProfilePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create installs.ini
	installsContent := `[ABC123]
Default=Profiles/installs-profile
Locked=1
`
	err := os.WriteFile(filepath.Join(tmpDir, "installs.ini"), []byte(installsContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfilePath(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/installs-profile"), profilePath)
}

func TestGetProfilePathFallsBackToProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// No installs.ini, only profiles.ini
	profilesContent := `[Profile0]
Name=fallback-profile
IsRelative=1
Path=Profiles/fallback-profile
Default=1
`
	err := os.WriteFile(filepath.Join(tmpDir, "profiles.ini"), []byte(profilesContent), 0644)
	require.NoError(t, err)

	profilePath, err := getProfilePath(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "Profiles/fallback-profile"), profilePath)
}

func TestFirefoxConfigLoadWithConfiguredPath(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &FirefoxConfig{ProfilePath: tmpDir}
	err := cfg.Load()
	assert.NoError(t, err)
	assert.Equal(t, tmpDir, cfg.ProfilePath)
}

func TestFirefoxConfigLoadWithInvalidConfiguredPath(t *testing.T) {
	cfg := &FirefoxConfig{ProfilePath: "/nonexistent/path"}
	// Should not error, but will try to auto-detect or use default
	err := cfg.Load()
	assert.NoError(t, err)
}

func TestLinuxFirefoxPaths(t *testing.T) {
	home := "/home/testuser"
	paths := linuxFirefoxPaths(home)
	assert.Contains(t, paths, filepath.Join(home, ".mozilla/firefox"))
	assert.Contains(t, paths, filepath.Join(home, "snap/firefox/common/.mozilla/firefox"))
	assert.Contains(t, paths, filepath.Join(home, ".var/app/org.mozilla.firefox/.mozilla/firefox"))
	assert.Len(t, paths, 3)
}

func TestMacOSFirefoxPaths(t *testing.T) {
	home := "/Users/testuser"
	paths := macOSFirefoxPaths(home)
	assert.Contains(t, paths, filepath.Join(home, "Library", "Application Support", "Firefox"))
	assert.Len(t, paths, 1)
}
