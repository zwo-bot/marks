package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataDir(t *testing.T) {
	// Set XDG_DATA_HOME to temp dir for testing
	tmpDir := t.TempDir()
	oldVal := os.Getenv("XDG_DATA_HOME")
	t.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldVal)

	dataDir, err := DataDir()
	assert.NoError(t, err)
	assert.Contains(t, dataDir, "marks")

	// Verify directory was created
	_, err = os.Stat(dataDir)
	assert.NoError(t, err)
}
