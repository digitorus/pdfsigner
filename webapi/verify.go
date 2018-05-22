package webapi

import (
	"fmt"
	"io"
	"net/http"

	"bitbucket.org/digitorus/pdfsigner/queues/queue"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

// handleVerifySchedule add a new verification job to the verification queue
func (wa *WebAPI) handleVerifySchedule(w http.ResponseWriter, r *http.Request) error {
	// put job with specified signer
	mr, err := r.MultipartReader()
	if err != nil {
		return httpError(w, errors2.Wrap(err, "read multipart"), 500)
	}

	fileNames := map[string]string{}

	for {
		// get part
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return httpError(w, errors2.Wrap(err, "get multipart"), 500)
		}

		//save pdf file to tmp
		err = savePDFToTemp(p, fileNames)
		if err != nil {
			return httpError(w, errors2.Wrap(err, "save pdf to tmp"), 500)
		}
	}

	jobID, err := addVerifyJob(wa.queue, fileNames)
	if err != nil {
		return httpError(w, errors2.Wrap(err, "push tasks"), 500)

	}

	_, err = fmt.Fprint(w, jobID)
	if err != nil {
		return err
	}

	return nil
}

// handleVerifyCheck check current status of the job
func (wa *WebAPI) handleVerifyCheck(w http.ResponseWriter, r *http.Request) error {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	// get job from the queue
	job, err := wa.queue.GetJobByID(jobID)
	if err != nil {
		return httpError(w, err, 500)
	}

	// respond with json
	return respondJSON(w, job, http.StatusOK)
}

// addVerifyJob adds verification job to the verification queue
func addVerifyJob(qs *queue.Queue, fileNames map[string]string) (string, error) {
	totalTasks := len(fileNames)

	// add job
	jobID := qs.AddVerifyJob()

	// determine priority
	priority := determinePriority(totalTasks)

	// add tasks
	for originalFileName, fileName := range fileNames {
		_, err := qs.AddTask(queue.VerificationUnitName, jobID, originalFileName, fileName, "", priority)
		if err != nil {
			return "", err
		}
	}

	return jobID, nil
}
