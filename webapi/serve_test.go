package webapi

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sync"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"io/ioutil"
)

func TestPut(t *testing.T) {
	addr := "localhost:3000"
	qsm := queued_sign.NewQSign()

	d := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType: 2,
			Approval: false,
		},
	}

	d.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")

	qsm.AddSigner("simple", d, 10)

	wa := NewWebAPI(addr, qsm, []string{
		"simple",
	})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		wa.Serve()
		wg.Done()
	}()
	go func() {
		qsm.Runner()
		wg.Done()
	}()

	req, err := newfileUploadRequest(
		"http://"+addr+"/put",
		map[string]string{
			"signer": "simple",
		},
		"file",
		"/Users/tim/go/src/bitbucket.org/digitorus/pdfsign/testfiles/testfile20.pdf")

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body))

	wg.Wait()
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
