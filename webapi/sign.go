package webapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"io/ioutil"

	"mime/multipart"

	"bitbucket.org/digitorus/pdfsigner/queues/queue"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

// handleSignScheduleResponse represents response for handleSignSchedule
type handleSignScheduleResponse struct {
	JobID string `json:"job_id"`
}

// handleSignSchedule adds a job to the queue
func (wa *WebAPI) handleSignSchedule(w http.ResponseWriter, r *http.Request) error {
	// put job with specified signer
	mr, err := r.MultipartReader()
	if err != nil {
		return httpError(w, errors2.Wrap(err, "read multipart"), http.StatusInternalServerError)
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
			return httpError(w, errors2.Wrap(err, "get multipart"), http.StatusInternalServerError)
		}

		//parse fields
		err = parseFields(p, &f)
		if err != nil {
			return httpError(w, errors2.Wrap(err, "parse fields"), http.StatusInternalServerError)
		}

		//save pdf file to tmp
		err = savePDFToTemp(p, &fileNames)
		if err != nil {
			return httpError(w, errors2.Wrap(err, "save pdf to tmp"), http.StatusInternalServerError)
		}
	}

	// check if at least one file was provided
	if len(fileNames) < 1 {
		return httpError(w, errors2.Wrap(errors.New("No files provided"), "validation"), http.StatusBadRequest)
	}

	// add job to the queue
	jobID, err := addSignJob(wa.queue, f, fileNames)
	if err != nil {
		return httpError(w, errors2.Wrap(err, "add tasks"), http.StatusInternalServerError)
	}

	// create response
	res := handleSignScheduleResponse{jobID}

	// set location
	w.Header().Set("Location", "/sign/"+jobID)

	// respond with json
	return respondJSON(w, res, http.StatusCreated)
}

func addSignJob(qs *queue.Queue, f fields, fileNames []string) (string, error) {
	if f.signerName == "" {
		return "", errors.New("signer name is required")
	}

	totalTasks := len(fileNames)

	jobID := qs.AddSignJob(f.signData)
	priority := determinePriority(totalTasks)

	err := qs.AddBatchPersistentTasks(f.signerName, jobID, fileNames, priority)
	if err != nil {
		return "", err
	}

	return jobID, nil
}

type job struct {
	ID string `json:"id"`
}
type task struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type JobStatus struct {
	Job   job    `json:"job"`
	Tasks []task `json:"tasks"`
}

func (wa *WebAPI) handleSignStatus(w http.ResponseWriter, r *http.Request) error {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	j, err := wa.queue.GetJobByID(jobID)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	status := r.URL.Query().Get("status")
	tasks, err := j.GetTasks(status)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	var responseTasks []task
	for _, t := range tasks {
		rt := task{ID: t.ID, Status: t.Status}
		responseTasks = append(responseTasks, rt)
	}

	jobStatus := JobStatus{Job: job{j.ID}, Tasks: responseTasks}

	return respondJSON(w, jobStatus, http.StatusOK)
}

func (wa *WebAPI) handleSignGetFile(w http.ResponseWriter, r *http.Request) error {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]
	taskID := vars["taskID"]

	// get file path
	filePath, err := wa.queue.GetCompletedTaskFilePath(jobID, taskID)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	// get file
	file, err := os.Open(filePath)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}
	defer file.Close()

	// get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	_, err = io.Copy(w, file)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	return nil
}

type fields struct {
	signerName string
	signData   queue.SignData
}

func parseFields(p *multipart.Part, f *fields) error {
	switch p.FormName() {
	case "signer", "name", "location", "reason", "contactInfo", "certType", "approval":
		//parse params
		slurp, err := ioutil.ReadAll(p)
		if err != nil {
			return nil
		}

		// get field content
		str := string(slurp)

		switch p.FormName() {
		case "signer":
			f.signerName = str
		case "name":
			f.signData.Name = str
		case "location":
			f.signData.Location = str
		case "reason":
			f.signData.Reason = str
		case "contactInfo":
			f.signData.ContactInfo = str
		case "certType":
			i, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			f.signData.CertType = uint32(i)
		case "approval":
			b, err := strconv.ParseBool(str)
			if err != nil {
				return err
			}
			f.signData.Approval = b
		}
	}

	return nil
}

// handleSignDelete removes job from the queue
func (wa *WebAPI) handleSignDelete(w http.ResponseWriter, r *http.Request) error {
	// get job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	// delete job by id
	err := wa.queue.DeleteJob(jobID)
	if err != nil {
		return httpError(w, errors2.Wrap(err, "couldn't delete job"), http.StatusInternalServerError)
	}

	// respond with ok
	w.WriteHeader(http.StatusOK)
	return nil
}
