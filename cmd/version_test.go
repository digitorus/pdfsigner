package cmd

import (
	"testing"

	"github.com/digitorus/pdfsigner/version"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	// Save original version and restore after tests
	originalVersion := ver
	defer func() { ver = originalVersion }()

	// Set test version data
	ver = version.Version{
		Version:   "1.0.0-test",
		BuildDate: "2023-01-01",
		GitCommit: "abc123",
		GitBranch: "main",
	}

	t.Run("Version command should display version information", func(t *testing.T) {
		output, err := executeCommand(t, "version")

		assert.NoError(t, err)
		assert.Contains(t, output, "Version 1.0.0-test")
		assert.Contains(t, output, "BuildDate 2023-01-01")
		assert.Contains(t, output, "GitCommit abc123")
		assert.Contains(t, output, "GitBranch main")
	})
}
