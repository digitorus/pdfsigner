package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyCommand(t *testing.T) {
	t.Run("Verify command should verify a signed PDF", func(t *testing.T) {
		args := []string{"verify", signedPDF}

		output, err := executeCommandWithArgs(t, RootCmd, args)
		t.Log("output:", output)

		assert.NoError(t, err)
		assert.Contains(t, output, "verified successfully", output)
	})

	t.Run("Verify command should fail without file arguments", func(t *testing.T) {
		args := []string{"verify"}

		output, err := executeCommandWithArgs(t, RootCmd, args)

		assert.Error(t, err)
		assert.Contains(t, output, "no files provided")
	})
}
