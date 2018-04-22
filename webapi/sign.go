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
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

// handleSignScheduleResponse represents response for handleSignSchedule
type handleSignScheduleResponse struct {
	JobID string `json:"job_id"`
}

// handleSignSchedule adds a job to the queue
func (wa *WebAPI) handleSignSchedule(w http.ResponseWriter, r *http.Request) {
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

	// add job to the queue
	jobID, err := addSignJob(wa.queue, f, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "add tasks"), 500)
		return
	}

	// create response
	res := handleSignScheduleResponse{jobID}

	// set location
	w.Header().Set("Location", "/sign/"+jobID)

	// respond with json
	respondJSON(w, res, http.StatusCreated)
}

func addSignJob(qs *queue.Queue, f fields, fileNames []string) (string, error) {
	if f.signerName == "" {
		return "", errors.New("signer name is required")
	}

	totalTasks := len(fileNames)

	jobID := qs.AddJob(f.signData)
	priority := determinePriority(totalTasks)

	for _, fileName := range fileNames {
		_, err := qs.AddTask(f.signerName, jobID, fileName, fileName+"_signed.pdf", priority)
		if err != nil {
			return "", err
		}
	}

	return jobID, nil
}

type JobStatus struct {
	queue.Job
	Tasks []queue.Task `json:"tasks"`
}

func (wa *WebAPI) handleSignStatus(w http.ResponseWriter, r *http.Request) {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	job, err := wa.queue.GetJobByID(jobID)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	status := r.URL.Query().Get("status")
	tasks, err := job.GetTasks(status)
	if err != nil {
		httpError(w, err, 500)
	}

	jobStatus := JobStatus{job, tasks}

	respondJSON(w, jobStatus, http.StatusOK)
}

func (wa *WebAPI) handleSignGetFile(w http.ResponseWriter, r *http.Request) {
	// get tasks for job
	vars := mux.Vars(r)
	jobID := vars["jobID"]
	taskID := vars["taskID"]

	// get file path
	filePath, err := wa.queue.GetCompletedTaskFilePath(jobID, taskID)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// get file
	file, err := os.Open(filePath)
	if err != nil {
		httpError(w, err, 500)
		return
	}
	defer file.Close()

	// get file info
	fileInfo, err := file.Stat()
	if err != nil {
		httpError(w, err, 500)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	_, err = io.Copy(w, file)
	if err != nil {
		httpError(w, err, 500)
		return
	}
}

type fields struct {
	signerName string
	signData   signer.SignData
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
			f.signData.Signature.Info.Name = str
		case "location":
			f.signData.Signature.Info.Location = str
		case "reason":
			f.signData.Signature.Info.Reason = str
		case "contactInfo":
			f.signData.Signature.Info.ContactInfo = str
		case "certType":
			i, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			f.signData.Signature.CertType = uint32(i)
		case "approval":
			b, err := strconv.ParseBool(str)
			if err != nil {
				return err
			}
			f.signData.Signature.Approval = b
		}
	}

	return nil
}

// handleSignDelete removes job from the queue
func (wa *WebAPI) handleSignDelete(w http.ResponseWriter, r *http.Request) {
	// get job
	vars := mux.Vars(r)
	jobID := vars["jobID"]

	// delete job by id
	wa.queue.DeleteJob(jobID)

	// respond with ok
	w.WriteHeader(http.StatusOK)
}
