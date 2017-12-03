package files

import (
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/spf13/viper"
)

func storeTempFile(file io.Reader) (string, error) {
	// TODO: Should we encrypt temporary files?
	tmpFile, err := ioutil.TempFile("", "pdf")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

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

//signer.SignFile(fullFilePath, outputFolder, signData)

func SignFilesByPatterns(filePatterns []string, signData signer.SignData) {
	files, err := findFilesByPatterns(filePatterns)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if err := signer.SignFile(f, viper.GetString("out"), signData); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Signed PDF written to " + viper.GetString("out"))
}
