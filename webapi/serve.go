package webapi

import (
	"encoding/json"
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
	r              *mux.Router
	addr           string
	qSign          queued_sign.QSign
	allowedSigners []string
}

func NewWebAPI(addr string, qs queued_sign.QSign, allowedSigners []string) *WebAPI {
	wa := WebAPI{
		addr:           addr,
		qSign:          qs,
		allowedSigners: allowedSigners,
		r:              mux.NewRouter(),
	}

	wa.r.HandleFunc("/put", wa.handlePut).Methods("POST")
	wa.r.HandleFunc("/check/{sessionID}", wa.handleCheckBySessionID).Methods("GET")
	wa.r.HandleFunc("/get/", wa.handleGetBySessionID).Methods("GET")

	return &wa
}

func (wa *WebAPI) Serve() {
	s := &http.Server{
		Addr:           wa.addr,
		Handler:        wa.r,
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
		err = savePDFToTemp(p, &fileNames)
		if err != nil {
			httpError(w, errors2.Wrap(err, "save pdf to tmp"), 500)
			return
		}
	}

	sessionID, err := pushJobs(wa.qSign, f, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "push jobs"), 500)
		return
	}

	_, err = fmt.Fprint(w, sessionID)
	if err != nil {
		log.Println(err)
	}
}

func (wa *WebAPI) handleCheckBySessionID(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]
	log.Println(sessionId)

	sess, err := wa.qSign.GetSessionByID(sessionId)
	if err != nil {
		httpError(w, err, 500)
	}

	// respond with json
	j, err := json.Marshal(sess)
	if err != nil {
		httpError(w, err, 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)

}

func (wa *WebAPI) handleGetBySessionID(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]
	jobs, err := wa.qSign.GetSessionCompletedJobs(sessionId)
	if err != nil {
		httpError(w, err, 500)
	}

	// respond with json
	j, err := json.Marshal(jobs)
	if err != nil {
		httpError(w, err, 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

type fields struct {
	signerName string
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

func savePDFToTemp(p *multipart.Part, fileNames *[]string) error {
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

		*fileNames = append(*fileNames, f.Name())
	}

	return nil
}

func pushJobs(qs queued_sign.QSign, f fields, fileNames []string) (string, error) {
	if f.signerName == "" {
		return "", errors.New("signer name is required")
	}

	totalJobs := len(fileNames)
	sessionID := qs.NewSession(totalJobs)
	priority := determinePriority(totalJobs)

	for _, fileName := range fileNames {
		_, err := qs.PushJob(f.signerName, sessionID, fileName, fileName+"_signed.pdf", priority)
		if err != nil {
			return "", err
		}
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
