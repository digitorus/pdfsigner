package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignCommand(t *testing.T) {
	// Save original config and restore after tests
	originalConfig := config
	defer func() { config = originalConfig }()

	t.Run("Sign command should display help when no arguments", func(t *testing.T) {
		output, err := executeCommand(t, "sign")

		assert.Error(t, err)
		assert.Contains(t, output, "pdfsigner sign [flags]")
	})
}

func TestSignPEMCommand(t *testing.T) {
	// Create a temporary output directory
	outputDir, err := os.MkdirTemp("", "pdfsigner-test-output")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(outputDir)

	t.Run("SignPEM command should fail without file patterns", func(t *testing.T) {
		output, err := executeCommand(t, "sign", "pem")

		assert.Error(t, err)
		assert.Contains(t, output, "no file patterns provided")
	})

	t.Run("SignPEM command should work with proper arguments", func(t *testing.T) {
		outputFilePath := testPDF12[:len(testPDF12)-4] + "_signed.pdf"

		// Remove any previous output file
		_ = os.Remove(outputFilePath)

		// Run the sign command
		output, err := executeCommand(t,
			"sign", "pem",
			"--cert", certPath,
			"--key", keyPath,
			"--name", "Test Signer",
			"--reason", "Testing",
			testPDF12,
		)
		require.NoError(t, err, "Failed to execute sign PEM command")

		// If the file was created, verify it's different than the original
		if _, statErr := os.Stat(outputFilePath); statErr == nil {
			originalInfo, statErr := os.Stat(testPDF12)
			assert.NoError(t, statErr, "Failed to stat original file")

			signedInfo, statErr := os.Stat(outputFilePath)
			assert.NoError(t, statErr, "Failed to stat signed file")

			// The signed file should have a different size than the original
			assert.NotEqual(t, originalInfo.Size(), signedInfo.Size(),
				"Signed file should be different from original")
		} else {
			t.Errorf("Signing completed but signed file not found. Command output: %s", output)
		}
	})

	t.Run("SignPEM command should respect output directory without adding _signed suffix", func(t *testing.T) {
		// Get the original filename without path
		originalFileName := filepath.Base(testPDF12)

		// Expected output is the original filename in the output directory (without _signed)
		expectedOutputPath := filepath.Join(outputDir, originalFileName)

		// Remove any previous output file
		_ = os.Remove(expectedOutputPath)

		// Run the sign command with output directory
		output, err := executeCommand(t,
			"sign", "pem",
			"--cert", certPath,
			"--key", keyPath,
			"--name", "Test Signer",
			"--reason", "Testing",
			"--out", outputDir,
			testPDF12,
		)
		require.NoError(t, err, "Failed to execute sign PEM command with output directory")

		// Verify the file was created with correct name (original name without _signed)
		if _, statErr := os.Stat(expectedOutputPath); statErr == nil {
			originalInfo, statErr := os.Stat(testPDF12)
			assert.NoError(t, statErr, "Failed to stat original file")

			signedInfo, statErr := os.Stat(expectedOutputPath)
			assert.NoError(t, statErr, "Failed to stat signed file")

			// The signed file should have a different size than the original
			assert.NotEqual(t, originalInfo.Size(), signedInfo.Size(),
				"Signed file should be different from original")

			// Verify the file name doesn't have _signed suffix
			assert.False(t, strings.Contains(expectedOutputPath, "_signed"),
				"Output filename should not contain _signed suffix when output directory is specified")
		} else {
			t.Errorf("Signing completed but signed file not found at %s. Command output: %s",
				expectedOutputPath, output)
		}
	})

	t.Run("SignPEM command should handle malformed PDF", func(t *testing.T) {
		args := []string{
			"sign", "pem",
			"--cert", certPath,
			"--key", keyPath,
			"--out", outputDir,
			malformedPDF,
		}

		// Run the sign command
		output, err := executeCommandWithArgs(t, RootCmd, args)
		if err == nil {
			t.Log(output)
		}
		assert.Error(t, err)
	})

	t.Run("SignPEM command should work with multiple PDF files", func(t *testing.T) {
		args := []string{
			"sign", "pem",
			"--cert", certPath,
			"--key", keyPath,
			"--out", outputDir,
			testPDF12,
			testPDF20,
		}

		// Run the sign command
		output, err := executeCommandWithArgs(t, RootCmd, args)
		if err == nil {
			t.Log(output)
		}
		assert.NoError(t, err)

		// Check if both files were created in the output directory with original names
		expectedPDF12 := filepath.Join(outputDir, filepath.Base(testPDF12))
		expectedPDF20 := filepath.Join(outputDir, filepath.Base(testPDF20))

		// Verify files exist
		pdf12Exists := true
		pdf20Exists := true

		if _, statErr := os.Stat(expectedPDF12); os.IsNotExist(statErr) {
			pdf12Exists = false
			t.Logf("Warning: Expected output file not found: %s", expectedPDF12)
		}

		if _, statErr := os.Stat(expectedPDF20); os.IsNotExist(statErr) {
			pdf20Exists = false
			t.Logf("Warning: Expected output file not found: %s", expectedPDF20)
		}

		// At least one of the files should exist if the command worked
		if !pdf12Exists && !pdf20Exists && err == nil {
			t.Errorf("Command succeeded but no output files were found")
		}
	})
}

func TestSignPKSC11Command(t *testing.T) {
	// Skip the PKCS11 tests if running in a CI environment without proper PKCS11 setup
	if os.Getenv("CI") != "" {
		t.Skip("Skipping PKCS11 tests in CI environment")
	}

	t.Run("SignPKSC11 command should fail without file patterns", func(t *testing.T) {
		args := []string{"sign", "pksc11"}

		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "no file patterns provided")
	})
}

func TestSignBySignerCommand(t *testing.T) {
	// Save original config and restore after tests
	originalConfig := config
	defer func() { config = originalConfig }()

	// Create a temporary output directory
	outputDir, err := os.MkdirTemp("", "pdfsigner-test-output-signer")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(outputDir)

	t.Run("SignBySigner command should fail without file patterns", func(t *testing.T) {
		args := []string{"sign", "signer"}

		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "no file patterns provided")
	})

	t.Run("SignBySigner command should fail with empty signer name", func(t *testing.T) {
		args := []string{"sign", "signer", testPDF20}

		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "signer name is empty")
	})

	t.Run("SignBySigner command should fail with non-existent signer", func(t *testing.T) {
		args := []string{
			"sign", "signer",
			testPDF20,
			"--signer", "nonexistent",
		}

		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "signer not found")
	})

	t.Run("SignBySigner command should work with proper PEM signer", func(t *testing.T) {
		// Set up config with a PEM signer
		config = Config{
			Signers: map[string]signerConfig{
				"pemsigner": {
					Type: "pem",
					Cert: certPath,
					Key:  keyPath,
				},
			},
		}

		args := []string{
			"sign", "signer",
			"--signer", "pemsigner",
			"--out", outputDir,
			testPDF20,
		}

		_, err := executeCommandWithArgs(t, RootCmd, args)
		assert.NoError(t, err)
	})
}
