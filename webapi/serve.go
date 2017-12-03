package webapi

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"io/ioutil"
	"path"

	"mime/multipart"

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

type WebAPI struct {
	addr           string
	qSign          queued_sign.QSign
	allowedSigners []string
}

func NewWebAPI(addr string, qs queued_sign.QSign, allowedSigners []string) *WebAPI {
	return &WebAPI{
		addr:           addr,
		qSign:          qs,
		allowedSigners: allowedSigners,
	}
}

func (wa *WebAPI) Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/put", wa.handlePut).Methods("POST")
	r.HandleFunc("/get", wa.handleGet).Methods("GET")
	r.HandleFunc("/check", wa.handleCheck).Methods("GET")

	s := &http.Server{
		Addr:           wa.addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}

func (wa *WebAPI) handlePut(w http.ResponseWriter, r *http.Request) {
	// put job with specified signer
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var f fields
	var fileNames []string

	for {
		// get part
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			httpError(w, errors2.Wrap(err, "get multipart"), 500)
			return
		}

		//parse fields
		err = parseFields(p, &f)
		if err != nil {
			httpError(w, errors2.Wrap(err, "parse fields"), 500)
			return
		}

		//save pdf file to tmp
		err = savePDFToTemp(p, fileNames)
		if err != nil {
			httpError(w, errors2.Wrap(err, "save pdf to tmp"), 500)
			return
		}
	}

	sessionID, err := pushJobs(wa.qSign, f, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "push jobs"), 500)
	}

	_, err = fmt.Fprint(w, sessionID)
	if err != nil {
		log.Println(err)
	}
}

type fields struct {
	signerName string
}

func httpError(w http.ResponseWriter, err error, code int) {
	fmt.Printf("%v", err)
	http.Error(w, err.Error(), code)
}

func parseFields(p *multipart.Part, f *fields) error {
	//parse params
	slurp, err := ioutil.ReadAll(p)
	if err != nil {
		return nil
	}

	switch p.FormName() {
	case "signer":
		f.signerName = string(slurp)
	case "other":
	}

	return nil
}

func savePDFToTemp(p *multipart.Part, fileNames []string) error {

	// parse pdf
	if path.Ext(p.FileName()) == ".pdf" {
		f, err := ioutil.TempFile("", "pdfsigner_cache")
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, p)
		if err != nil {
			return err
		}

		fileNames = append(fileNames, f.Name())
	}

	return nil
}

func pushJobs(m queued_sign.QSign, f fields, fileNames []string) (string, error) {

	if f.signerName == "" {
		return "", errors.New("signer with this name is not setup for this api")
	}

	sessionID := m.NewSession()

	priority := determinePriority(len(fileNames))

	for _, fileName := range fileNames {
		m.PushJob(sessionID, f.signerName, fileName, fileName+"_signed.pdf", priority)
	}

	return sessionID, nil
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

func (wa *WebAPI) handleCheck(w http.ResponseWriter, r *http.Request) {
}
func (wa *WebAPI) handleGet(w http.ResponseWriter, r *http.Request) {
}

func dumpRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%q", dump)
}
