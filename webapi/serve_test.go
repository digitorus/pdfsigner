package webapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/queues/queue"
	"github.com/digitorus/pdfsigner/signer"
	"github.com/digitorus/pdfsigner/version"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type filePart struct {
	fieldName string
	path      string
}

var (
	wa      *WebAPI
	q       *queue.Queue
	proto   = "http://"
	addr    = "localhost:3000"
	baseURL = proto + addr
)

// TestMain setup / tear down before tests.
func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

// runTest initializes the environment.
func runTest(m *testing.M) int {
	log.SetOutput(io.Discard)

	err := license.Initialize([]byte(license.TestLicense))
	if err != nil {
		log.Fatal(err)
	}

	// create new queue
	q = queue.NewQueue()

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
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
	}
	signData.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")
	q.AddSignUnit("simple", signData)
	q.AddVerifyUnit()
	q.StartProcessor()

	// create web api
	wa = NewWebAPI(addr, q, []string{
		"simple",
	}, version.Version{Version: "0.1"},
		true,
	)

	return m.Run()
}

func TestSignFlow(t *testing.T) {
	// test upload
	// create file parts
	fileParts := []filePart{
		{"testfile1", "../testfiles/testfile12.pdf"},
		{"testfile2", "../testfiles/testfile12.pdf"},
		{"testfile3-mal", "../testfiles/malformed.pdf"},
	}
	// create multipart request
	r, err := newMultipleFilesUploadRequest(
		baseURL+"/sign",
		map[string]string{
			"signer":      "simple",
			"name":        "My Name",
			"location":    "My Location",
			"reason":      "My Reason",
			"contactInfo": "My ContactInfo",
			"certType":    "1",
			"docMDP":      "1",
		}, fileParts)
	if err != nil {
		t.Fatal(err)
	}

	// create recorder
	w := httptest.NewRecorder()
	// make request
	wa.r.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	// get job id
	var scheduleResponse hanldeScheduleResponse
	if err := json.NewDecoder(w.Body).Decode(&scheduleResponse); err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, scheduleResponse.JobID)
	assert.Equal(t, "/sign/"+scheduleResponse.JobID, w.Header().Get("Location"), "location is not set")

	// wait for signing files
	time.Sleep(2 * time.Second)

	// test status
	r = httptest.NewRequest(http.MethodGet, baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var jobStatus jobStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&jobStatus); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, jobStatus.Tasks, 3)

	var completedTasks int

	for _, task := range jobStatus.Tasks {
		// work with failed malformed task
		if task.Status != queue.StatusCompleted {
			assert.NotEmpty(t, task.Error)

			continue
		}

		// happy path
		assert.Equal(t, "testfile12.pdf", task.OriginalFileName)
		// test get completed task
		r = httptest.NewRequest(http.MethodGet, baseURL+"/sign/"+scheduleResponse.JobID+"/"+task.ID+"/download", nil)
		w = httptest.NewRecorder()
		wa.r.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Len(t, w.Body.Bytes(), 20651)

		completedTasks += 1
	}

	// check amount of success
	assert.Equal(t, 2, completedTasks)

	// test delete job
	r = httptest.NewRequest(http.MethodDelete, baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	r = httptest.NewRequest(http.MethodGet, baseURL+"/sign/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)

	assert.NotEqual(t, http.StatusOK, w.Code, w.Body.String(), "not removed")

	// test get version
	r = httptest.NewRequest(http.MethodGet, baseURL+"/version", nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var ver version.Version
	if err := json.NewDecoder(w.Body).Decode(&ver); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "0.1", ver.Version)
}

func TestVerifyFlow(t *testing.T) {
	// create file parts
	fileParts := []filePart{
		{"testfile1", "../testfiles/testfile12_signed.pdf"},
	}
	// create multipart request
	r, err := newMultipleFilesUploadRequest(
		baseURL+"/verify",
		map[string]string{}, fileParts)
	if err != nil {
		t.Fatal(err)
	}

	// create recorder
	w := httptest.NewRecorder()
	// make request
	wa.r.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status not ok: %v", w.Body.String())
	}

	// get job id
	var scheduleResponse hanldeScheduleResponse
	if err := json.NewDecoder(w.Body).Decode(&scheduleResponse); err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, scheduleResponse.JobID)
	assert.Equal(t, "/verify/"+scheduleResponse.JobID, w.Header().Get("Location"), "location is not set")

	// wait for signing files
	time.Sleep(2 * time.Second)

	// test status
	r = httptest.NewRequest(http.MethodGet, baseURL+"/verify/"+scheduleResponse.JobID, nil)
	w = httptest.NewRecorder()
	wa.r.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var jobStatus jobStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&jobStatus); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, jobStatus.Tasks, 1)

	var completedTasks int

	for _, task := range jobStatus.Tasks {
		// work with failed malformed task
		if task.Status != queue.StatusCompleted {
			assert.NotEmpty(t, task.Error)

			continue
		}

		// happy path
		assert.Equal(t, "testfile12_signed.pdf", task.OriginalFileName)
		// test get completed task
		r = httptest.NewRequest(http.MethodGet, baseURL+"/verify/"+scheduleResponse.JobID+"/info/"+task.ID, nil)
		w = httptest.NewRecorder()
		wa.r.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		completedTasks += 1
	}

	// check amount of success
	assert.Equal(t, 1, completedTasks)
}

// Creates a new multiple files upload http request with optional extra params.
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

	req, err := http.NewRequest(http.MethodPost, uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}
