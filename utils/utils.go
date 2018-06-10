package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetRunFileFolder() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir, nil
}

func IsTestEnvironment() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

func GetGoPath() string {
	return os.Getenv("GOPATH")
}
