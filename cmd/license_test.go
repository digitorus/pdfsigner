package cmd

import (
	"path/filepath"
	"testing"

	"github.com/digitorus/pdfsigner/license"
	"github.com/stretchr/testify/assert"
)

func TestLicenseCommands(t *testing.T) {
	// Save original license data and restore after tests
	originalLicenseData := license.LD
	defer func() { license.LD = originalLicenseData }()

	t.Run("License command should show help when no args provided", func(t *testing.T) {
		output, err := executeCommand(t, "license")

		assert.NoError(t, err)
		assert.Contains(t, output, "Check and update license")
	})

	t.Run("License info command should run without error", func(t *testing.T) {
		output, err := executeCommand(t, "license", "info")

		assert.NoError(t, err)
		assert.Contains(t, output, "Licensed to ")
	})

	// t.Run("License setup command should handle invalid license", func(t *testing.T) {
	// 	// Create temp file with invalid license
	// 	tmpFile := filepath.Join(t.TempDir(), "invalid_license.txt")
	// 	err := os.WriteFile(tmpFile, []byte("invalid-license"), 0600)
	// 	assert.NoError(t, err)

	// 	// Execute command with invalid license file
	// 	_, err = executeCommand(t, "license", "setup", "--license-path", tmpFile)

	// 	// Should error with invalid license
	// 	assert.Error(t, err)
	// })

	t.Run("License setup with non-existent file should fail", func(t *testing.T) {
		nonExistentFile := filepath.Join(t.TempDir(), "non_existent_license.txt")

		_, err := executeCommand(t, "license", "setup", "--license-path", nonExistentFile)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read license file")
	})
}
