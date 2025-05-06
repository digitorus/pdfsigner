package files

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/signer"
	"github.com/pkg/errors"
)

// findFilesByPatterns finds all files matched the patterns.
func findFilesByPatterns(patterns []string) (matchedFiles []string, err error) {
	for _, f := range patterns {
		m, err := filepath.Glob(f)
		if err != nil {
			return matchedFiles, err
		}

		matchedFiles = append(matchedFiles, m...)
	}

	return matchedFiles, err
}

// SignFilesByPatterns signs files by matched patterns and stores them in the specified output directory
// or with _signed.pdf suffix in the same directory if no output directory is provided.
func SignFilesByPatterns(filePatterns []string, signData signer.SignData, validateSignature bool, outputDirectory string) error {
	// get files
	files, err := findFilesByPatterns(filePatterns)
	if err != nil {
		return fmt.Errorf("failed to find files by patterns: %w", err)
	}

	for _, f := range files {
		// get file name and extension
		dir, fileName := path.Split(f)
		fileNameArr := strings.Split(fileName, path.Ext(fileName))
		fileNameArr = fileNameArr[:len(fileNameArr)-1]
		fileNameNoExt := strings.Join(fileNameArr, "")
		ext := path.Ext(fileName)

		// generate signed file path based on output directory or original location
		var signedFilePath string
		if outputDirectory != "" {
			// When output directory is specified, use original filename without "_signed" suffix
			signedFilePath = path.Join(outputDirectory, fileName)
		} else {
			// When no output directory, append "_signed" suffix (original behavior)
			signedFilePath = path.Join(dir, fileNameNoExt+"_signed"+ext)
		}

		// sign file
		if err := signer.SignFile(f, signedFilePath, signData, validateSignature); err != nil {
			return errors.Wrap(err, "failed to sign file: "+fileName)
		}
	}

	err = license.LD.SaveLimitState()
	if err != nil {
		return fmt.Errorf("failed to save license limit state: %w", err)
	}

	return nil
}
