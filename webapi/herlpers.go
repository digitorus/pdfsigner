package webapi

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"

	"github.com/digitorus/pdfsigner/queues/priority_queue"
)

// savePDFToTemp saves pdf files to temporary folder
// TODO: check if the encryption is needed
func savePDFToTemp(p *multipart.Part, fileNames map[string]string) error {
	// return error if the provided file has not supported extension
	ext := path.Ext(p.FileName())
	if ext != "" && ext != ".pdf" {
		return fmt.Errorf("not supported file: %s", p.FileName())
	}

	// parse pdf
	if ext == ".pdf" {
		f, err := os.CreateTemp("", "pdfsigner_cache")
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

		fileNames[f.Name()] = p.FileName()
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

// dumpRequest dumps request, only for debugging
// func dumpRequest(r *http.Request) {
// 	dump, err := httputil.DumpRequest(r, true)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Printf("%q", dump)
// }
