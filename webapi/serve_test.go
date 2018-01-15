package webapi

import (
	"bytes"
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
)

type filePart struct {
	fieldName string
	path      string
}

var (
	wa        *WebAPI
	qs        *queued_sign.QSign
	sessionID string
	proto     = "http://"
	addr      = "localhost:3000"
	baseURL   = proto + addr
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

func TestUploadCheckDownload(t *testing.T) {
	// test upload
	//create file parts
	fileParts := []filePart{
		{"testfile1", "../testfiles/testfile20.pdf"},
		{"testfile2", "../testfiles/testfile20.pdf"},
	}
	// create multipart request
	r, err := newMultipleFilesUploadRequest(
		baseURL+"/put",
		map[string]string{
			"signer": "simple",
		}, fileParts)
	if err != nil {
		t.Fatal(err)
	}
	// create recorder
	w := httptest.NewRecorder()

	// make request
	wa.r.ServeHTTP(w, r)
	//
	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}
	// get session id
	sessionID := w.Body.String()
	if sessionID == "" {
		t.Fatal("not received sessionID")
	}

	// wait for signing files
	time.Sleep(1 * time.Second)

	// test check
	r = httptest.NewRequest("GET", baseURL+"/check-session/"+sessionID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	var session queued_sign.Session
	if err := json.NewDecoder(w.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, session.IsCompleted)
	assert.Equal(t, 2, len(session.CompletedJobs))
	assert.Equal(t, "", session.CompletedJobs[0].Error)
	assert.Equal(t, "", session.CompletedJobs[1].Error)

	// test get completed jobs
	r = httptest.NewRequest("GET", baseURL+"/get-file/"+sessionID+"/"+session.CompletedJobs[0].ID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	if len(w.Body.Bytes()) != 9001 {
		t.Fatalf("file size is not correct")
	}
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
