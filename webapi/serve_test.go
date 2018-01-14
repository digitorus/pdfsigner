package webapi

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/signer"
)

type filePart struct {
	fieldName string
	path      string
}

var (
	wa      *WebAPI
	qs      queued_sign.QSign
	proto   = "http://"
	addr    = "localhost:3000"
	baseURL = proto + addr
)

// TestMain setup / tear down before tests
func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

// runTest initializes the environment
func runTest(m *testing.M) int {
	// create new QSign
	qs = queued_sign.NewQSign()

	// create signer
	signData := signer.SignData{
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
	signData.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")
	qs.AddSigner("simple", signData, 10)
	qs.Runner()

	// create web api
	wa = NewWebAPI(addr, qs, []string{
		"simple",
	})

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	return m.Run()
}

func TestPut(t *testing.T) {
	// upload pdf files
	fileParts := []filePart{
		{"testfile1", "../testfiles/testfile20.pdf"},
		{"testfile2", "../testfiles/testfile20.pdf"},
	}
	r, err := newMultipleFilesUploadRequest(
		baseURL+"/put",
		map[string]string{
			"signer": "simple",
		}, fileParts)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)

	sessionID := w.Body.String()
	if sessionID == "" {
		t.Fatal("not received sessionID")
	}
	time.Sleep(5 * time.Second)
	r = httptest.NewRequest("GET", baseURL+"/check/"+sessionID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)

	log.Println(w.Body.String())
}

// Creates a new multiple files upload http request with optional extra params
func newMultipleFilesUploadRequest(uri string, params map[string]string, fileParts []filePart) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	for _, f := range fileParts {
		file, err := os.Open(f.path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(f.fieldName, filepath.Base(f.path))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
