package webapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"bitbucket.org/digitorus/pdfsigner/queued_verify"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

func (wa *WebAPI) handleVerifySchedule(w http.ResponseWriter, r *http.Request) {
	// put job with specified signer
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

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

		//save pdf file to tmp
		err = savePDFToTemp(p, &fileNames)
		if err != nil {
			httpError(w, errors2.Wrap(err, "save pdf to tmp"), 500)
			return
		}
	}

	sessionID, err := pushVerifyJob(wa.qVerify, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "push jobs"), 500)
		return
	}

	_, err = fmt.Fprint(w, sessionID)
	if err != nil {
		log.Println(err)
	}
}

func (wa *WebAPI) handleVerifyCheck(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]

	sess, err := wa.qVerify.GetSessionByID(sessionId)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// respond with json
	j, err := json.Marshal(sess)
	if err != nil {
		httpError(w, err, 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func pushVerifyJob(qs *queued_verify.QVerify, fileNames []string) (string, error) {
	totalJobs := len(fileNames)

	sessionID := qs.NewSession(totalJobs)
	priority := determinePriority(totalJobs)

	for _, fileName := range fileNames {
		_, err := qs.PushJob(sessionID, fileName, priority)
		if err != nil {
			return "", err
		}
	}

	return sessionID, nil
}
