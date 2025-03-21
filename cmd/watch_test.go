package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchCommand(t *testing.T) {
	t.Run("Watch command should show help when no args provided", func(t *testing.T) {
		output, err := executeCommand(t, "watch")

		assert.NoError(t, err)
		assert.Contains(t, output, "Watch folder for new")
	})
}

func TestWatchStartWatch(t *testing.T) {
	// Create test input and output directories
	inputDir := filepath.Join(t.TempDir(), "watch-input")
	outputDir := filepath.Join(t.TempDir(), "watch-output")

	err := os.Mkdir(inputDir, 0755)
	require.NoError(t, err, "Failed to create input directory")

	err = os.Mkdir(outputDir, 0755)
	require.NoError(t, err, "Failed to create output directory")

	t.Run("startWatch should fail if input directory doesn't exist", func(t *testing.T) {
		inputDir := "/non/existent/directory"
		outputDir := filepath.Join(t.TempDir(), "watch-output")

		err = os.Mkdir(outputDir, 0755)
		require.NoError(t, err, "Failed to create output directory")

		args := []string{
			"watch", "pem",
			"--in", inputDir,
			"--out", outputDir,
		}

		// Without certificate and key, should fail
		_, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("startWatch should fail if output directory doesn't exist", func(t *testing.T) {
		inputDir := filepath.Join(t.TempDir(), "watch-input")
		outputDir := "/non/existent/directory"

		err := os.Mkdir(inputDir, 0755)
		require.NoError(t, err, "Failed to create input directory")

		args := []string{
			"watch", "pem",
			"--in", inputDir,
			"--out", outputDir,
		}

		// Without certificate and key, should fail
		_, err = executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestWatchCommands(t *testing.T) {
	// Create test input and output directories
	inputDir := filepath.Join(t.TempDir(), "watch-input")
	outputDir := filepath.Join(t.TempDir(), "watch-output")

	err := os.Mkdir(inputDir, 0755)
	require.NoError(t, err, "Failed to create input directory")

	err = os.Mkdir(outputDir, 0755)
	require.NoError(t, err, "Failed to create output directory")

	t.Run("Watch PEM command validation", func(t *testing.T) {
		args := []string{
			"watch", "pem",
			"--in", inputDir,
			"--out", outputDir,
		}

		// Without certificate and key, should fail
		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "failed to set PEM certificate data")
	})

	t.Run("Watch PKSC11 command validation", func(t *testing.T) {
		args := []string{
			"watch", "pksc11",
			"--in", inputDir,
			"--out", outputDir,
		}

		// Running without lib would fail in real execution but we'll just test command structure
		_, err := executeCommandWithArgs(t, RootCmd, args)
		assert.Error(t, err, "PKSC11 command should fail without proper parameters")
	})

	t.Run("Watch signer command validation", func(t *testing.T) {
		// Set a test signer in config
		originalConfig := config
		defer func() { config = originalConfig }()

		config = Config{
			Signers: map[string]signerConfig{
				"test-signer": {
					Type: "pem",
					Cert: certPath,
					Key:  keyPath,
				},
			},
		}

		args := []string{
			"watch", "signer",
			"--in", inputDir,
			"--out", outputDir,
			"--signerName", "test-signer",
		}

		// This should actually try to start the watch process
		_, err := executeCommandWithArgs(t, RootCmd, args)

		// The command structure is valid, but actual file watching would be started
		// We only care about validating the command structure for now
		assert.Error(t, err)
	})
}
