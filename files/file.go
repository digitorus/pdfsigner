package files

import (
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/signer"
)

// func storeTempFile(file io.Reader) (string, error) {
// 	// TODO: Should we encrypt temporary files?
// 	tmpFile, err := os.CreateTemp("", "pdf")
// 	if err != nil {
// 		return "", err
// 	}

// 	_, err = io.Copy(tmpFile, file)
// 	if err != nil {
// 		return "", err
// 	}
// 	return tmpFile.Name(), nil
// }

// findFilesByPatterns finds all files matched the patterns
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

// SignFilesByPatterns signs files by matched patterns and stores it inside the same folder with _signed.pdf suffix
func SignFilesByPatterns(filePatterns []string, signData signer.SignData, validateSignature bool) {
	// get files
	files, err := findFilesByPatterns(filePatterns)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		// generate signed file path
		dir, fileName := path.Split(f)
		fileNameArr := strings.Split(fileName, path.Ext(fileName))
		fileNameArr = fileNameArr[:len(fileNameArr)-1]
		fileNameNoExt := strings.Join(fileNameArr, "")
		signedFilePath := path.Join(dir, fileNameNoExt+"_signed"+path.Ext(fileName))

		// sign file
		if err := signer.SignFile(f, signedFilePath, signData, validateSignature); err != nil {
			log.Fatal(err)
		}
	}

	err = license.LD.SaveLimitState()
	if err != nil {
		log.Fatal(err)
	}
}
