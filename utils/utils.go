package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetRunFileFolder returns path to executable file
func GetRunFileFolder() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return dir, nil
}

// IsTestEnvironment checks if it runs inside the test
func IsTestEnvironment() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}
