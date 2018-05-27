package webapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"bitbucket.org/digitorus/pdfsigner/queues/queue"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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
		return httpError(w, errors.Wrap(err, "read multipart"), http.StatusInternalServerError)
	}

	var f fields
	fileNames := map[string]string{}

	for {
		// get part
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return httpError(w, errors.Wrap(err, "get multipart"), http.StatusBadRequest)
		}

		//parse fields
		err = parseFields(p, &f)
		if err != nil {
			return httpError(w, errors.Wrap(err, "parse fields"), http.StatusBadRequest)
		}

		//save pdf file to tmp
		err = savePDFToTemp(p, fileNames)
		if err != nil {
			return httpError(w, errors.Wrap(err, "save pdf to tmp"), http.StatusBadRequest)
		}
	}

	err = validate(f, fileNames)
	if err != nil {
		return httpError(w, errors.Wrap(err, "validation"), http.StatusBadRequest)
	}

	// add job to the queue
	jobID, err := addSignJob(wa.queue, f, fileNames)
	if err != nil {
		return httpError(w, errors.Wrap(err, "add tasks"), http.StatusBadRequest)
	}

	// create response
	res := handleSignScheduleResponse{jobID}

	// set location
	w.Header().Set("Location", "/sign/"+jobID)

	// respond with json
	return respondJSON(w, res, http.StatusCreated)
}

func validate(f fields, fileNames map[string]string) error {
	// check if at least one file was provided
	if len(fileNames) < 1 {
		return errors.New("no files provided")
	}

	if f.signerName == "" {
		return errors.New("signer name is not provided")
	}

	return nil
}

func addSignJob(qs *queue.Queue, f fields, fileNames map[string]string) (string, error) {

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
	ID               string `json:"id"`
	OriginalFileName string `json:"file_name"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
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
		return httpError(w, err, http.StatusBadRequest)
	}

	status := r.URL.Query().Get("status")
	tasks, err := j.GetTasks(status)
	if err != nil {
		return httpError(w, err, http.StatusBadRequest)
	}

	var responseTasks []task
	for _, t := range tasks {
		rt := task{ID: t.ID, Status: t.Status, OriginalFileName: t.OriginalFileName, Error: t.Error}
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
	completedTask, err := wa.queue.GetCompletedTask(jobID, taskID)
	if err != nil {
		return httpError(w, err, http.StatusBadRequest)
	}

	// get file
	file, err := os.Open(completedTask.OutputFilePath)
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
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, completedTask.OriginalFileName))
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
		return httpError(w, errors.Wrap(err, "couldn't delete job"), http.StatusBadRequest)
	}

	// respond with ok
	w.WriteHeader(http.StatusOK)
	return nil
}
