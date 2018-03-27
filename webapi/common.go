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

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
)

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

func determinePriority(totalTasks int) priority_queue.Priority {
	var priority priority_queue.Priority
	if totalTasks == 1 {
		priority = priority_queue.HighPriority
	} else {
		priority = priority_queue.MediumPriority
	}
	return priority
}

type httpErr struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func httpError(w http.ResponseWriter, err error, code int) {
	e := httpErr{Message: err.Error(), Code: code}
	// respond with json
	j, err := json.Marshal(e)
	if err != nil {
		httpError(w, err, 500)
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func dumpRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%q", dump)
}
