package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"path"

	"bitbucket.org/digitorus/pdfsigner/queues/priority_queue"
)

// savePDFToTemp saves pdf files to temporary folder
// TODO: check if the encryption is needed
func savePDFToTemp(p *multipart.Part, fileNames *[]string) error {
	// return error if the provided file has not supported extension
	ext := path.Ext(p.FileName())
	if ext != "" && ext != ".pdf" {
		return fmt.Errorf("not supported file: %s", p.FileName())
	}

	// parse pdf
	if ext == ".pdf" {
		f, err := ioutil.TempFile("", "pdfsigner_cache")
		if err != nil {
			return err
		}
		defer f.Close()

		written, err := io.Copy(f, p)
		if err != nil {
			return err
		}
		if written == 0 {
			return errors.New("written 0 bytes")
		}

		*fileNames = append(*fileNames, f.Name())
	}

	return nil
}

// determinePriority determines priority based on amount of the tasks needed to process
func determinePriority(totalTasks int) priority_queue.Priority {
	var priority priority_queue.Priority
	if totalTasks == 1 {
		priority = priority_queue.HighPriority
	} else {
		priority = priority_queue.MediumPriority
	}
	return priority
}

// httpErr represents the error response to the user
type httpErr struct {
	// Message represents error message
	Message string `json:"message"`
	// Code represents error code
	Code int `json:"code"`
}

// httpError writes to the response writer error and the code in json format
func httpError(w http.ResponseWriter, err error, code int) {
	e := httpErr{Message: err.Error(), Code: code}
	// respond with json
	respondJSON(w, e, code)
}

// respondJSON responds with json
func respondJSON(w http.ResponseWriter, data interface{}, code int) {
	// marshal data
	j, err := json.Marshal(data)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// set response code
	w.WriteHeader(code)

	// set content type
	w.Header().Set("Content-Type", "application/json")

	// respond with json
	w.Write(j)
}

// dumpRequest dumps request, only for debugging
func dumpRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%q", dump)
}
