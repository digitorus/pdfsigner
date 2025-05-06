package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Global test variables
var (
	testfilesDir string
	certPath     string
	keyPath      string
	testPDF12    string
	testPDF20    string
	malformedPDF string
	signedPDF    string
)

// TestMain sets up the test environment before any tests run
func TestMain(m *testing.M) {
	// Set up the testfiles directory path
	cwd, err := os.Getwd()
	if err != nil {
		panic("Failed to get current directory: " + err.Error())
	}

	// Use "../testfiles" as requested
	testfilesDir = filepath.Join(cwd, "../testfiles")

	// Set up common test file paths
	certPath = filepath.Join(testfilesDir, "test.crt")
	keyPath = filepath.Join(testfilesDir, "test.pem")
	testPDF12 = filepath.Join(testfilesDir, "testfile12.pdf")
	testPDF20 = filepath.Join(testfilesDir, "testfile20.pdf")
	malformedPDF = filepath.Join(testfilesDir, "malformed.pdf")
	signedPDF = filepath.Join(testfilesDir, "SampleSignedPDFDocument.pdf")

	// Verify test files existence before running any tests
	filesToCheck := []string{certPath, keyPath, testPDF12, testPDF20, malformedPDF, signedPDF}
	for _, file := range filesToCheck {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			panic("Required test file missing: " + file)
		}
	}

	// Run the tests
	os.Exit(m.Run())
}

// Simplified helper function to execute a command and capture its output
// Now always uses RootCmd implicitly
func executeCommand(t *testing.T, args ...string) (string, error) {
	// Create a fresh copy of the root command to avoid state between test runs
	cmd := RootCmd

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	t.Logf("Executing command: pdfsigner %s", strings.Join(args, " "))

	// Execute the command
	var err error
	assert.NotPanics(t, func() {
		err = cmd.Execute()
	})

	return buf.String(), err
}

// Keep executeCommandWithArgs for backward compatibility or special cases
func executeCommandWithArgs(t *testing.T, cmd *cobra.Command, args []string) (string, error) {
	if cmd == RootCmd {
		return executeCommand(t, args...)
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	t.Logf("Executing command: %s %s", cmd.CommandPath(), strings.Join(args, " "))

	// Execute the command
	var err error
	assert.NotPanics(t, func() {
		err = cmd.Execute()
	})

	return buf.String(), err
}
