package webapi

import (
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

func determinePriority(totalJobs int) priority_queue.Priority {
	var priority priority_queue.Priority
	if totalJobs == 1 {
		priority = priority_queue.HighPriority
	} else {
		priority = priority_queue.MediumPriority
	}
	return priority
}

func httpError(w http.ResponseWriter, err error, code int) {
	fmt.Printf("%v", err)
	http.Error(w, err.Error(), code)
}

func dumpRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%q", dump)
}
