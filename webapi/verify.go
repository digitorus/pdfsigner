package webapi

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"bitbucket.org/digitorus/pdfsigner/verify_queue"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

// handleVerifySchedule add a new verification job to the verification queue
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

	jobID, err := addVerifyJob(wa.qVerify, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "push tasks"), 500)
		return
	}

	_, err = fmt.Fprint(w, jobID)
	if err != nil {
		log.Println(err)
	}
}

// handleVerifyCheck check current status of the job
func (wa *WebAPI) handleVerifyCheck(w http.ResponseWriter, r *http.Request) {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	// get job from the queue
	job, err := wa.qVerify.GetJobByID(jobID)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// respond with json
	respondJSON(w, job, http.StatusOK)
}

// addVerifyJob adds verification job to the verification queue
func addVerifyJob(qs *verify_queue.QVerify, fileNames []string) (string, error) {
	totalTasks := len(fileNames)

	// add job
	jobID := qs.AddJob(totalTasks)

	// determine priority
	priority := determinePriority(totalTasks)

	// add tasks
	for _, fileName := range fileNames {
		_, err := qs.AddTask(jobID, fileName, priority)
		if err != nil {
			return "", err
		}
	}

	return jobID, nil
}
