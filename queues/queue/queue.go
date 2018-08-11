package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"

	"bitbucket.org/digitorus/pdfsigner/db"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsign/verify"
	"bitbucket.org/digitorus/pdfsigner/queues/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/rs/xid"
)

var (
	// StatusPending represents
	StatusPending = "Pending"
	// StatusFailed is a failed status
	StatusFailed = "Failed"
	// StatusCompleted is a completed status
	StatusCompleted = "Completed"
	// VerificationUnitName represents a task that should not be signed but verified
	VerificationUnitName = "VerificationUnitName"
)

// Queue represents sign queue
type Queue struct {
	// units represent all the units by name of the signer
	units map[string]*unit
	// jobs represents jobs by id of the job
	jobs map[string]*Job
	// db represents database container which used to save jobs
}

// unit represents queue unit which could be a signer or verifier
type unit struct {
	// name represents the name of the signer
	name string
	// pq represents priority queue used by the signer
	pq *priority_queue.PriorityQueue
	// isSigningUnit should be set to true if the unit is used for signing or false for verification
	isSigningUnit bool
	// signData represents sign data and it's used for signing unit
	signData signer.SignData
}

// Job represents a job for sign queue, stores tasks and sign data to override units initial sign data
type Job struct {
	// ID represents id of the job
	ID string `json:"id"`
	// TasksMap represents tasks added to the job
	TasksMap map[string]Task `json:"task_map"`
	// totalProcessedTasks represents total processed tasks of the job, incremented atomically
	TotalProcesedTasks uint32 `json:"total_procesed_tasks"`
	// JobSignConfig represents additional sign data added by request to override signer initial sign data
	SignConfig JobSignConfig `json:"sign_data"`
}

type JobSignConfig struct {
	// sign data
	Signer      string `json:"signer"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Reason      string `json:"reason"`
	ContactInfo string `json:"contact_info"`
	CertType    uint   `json:"cert_type"`
	DocMDPPerms uint   `json:"doc_mdp_perms"`
	// ValidateSignature allows to verify the job after it's being singed
	ValidateSignature bool `json:"verify_after_sign"`
}

// Task represents a single unit of work(file)
type Task struct {
	// ID represents id of the task
	ID string `json:"id"`
	// JobID represents id of the job task is assigned to
	JobID string `json:"job_id"`
	// OriginalFileName represents pdf file name
	OriginalFileName string `json:"original_file_name"`
	// InputFilePath represents path to the unprocessed file
	InputFilePath string `json:"input_file_path"`
	// OutputFilePath represents path to the processed file
	OutputFilePath string `json:"output_file_path"`
	// Status represents the status of the task. Pending, Failed, Completed.
	Status string `json:"status"`
	// VerificationData represents data of the verification
	VerificationData *verify.Response `json:"verification_data,omitempty"`
	// Error represents error if the task failed
	Error string `json:"error,omitempty"`
}

// GetTasks returns all the completed tasks if status is empty string,
// and only tasks with specific status if status is provided
func (j *Job) GetTasks(status string) ([]Task, error) {
	// determine status to search with
	switch status {
	case StatusCompleted:
	case StatusFailed:
	case StatusPending:
	case "":
	default:
		// fail if the status is not in the list
		return []Task{}, errors.New("status is not correct")
	}

	// find tasks by status
	var tasks []Task
	for _, t := range j.TasksMap {
		if status == "" || t.Status == status {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

// NewQueue creates new sign queue
func NewQueue() *Queue {
	return &Queue{
		units: make(map[string]*unit, 1),
		jobs:  make(map[string]*Job, 1),
	}
}

// addUnit adds unit to units map
func (q *Queue) addUnit(unitName string) *unit {
	// skip if already setup
	if _, exists := q.units[unitName]; exists {
		return nil
	}

	// create signer
	u := unit{
		name: unitName,
		pq:   priority_queue.New(10),
	}

	// assign signer to units map
	q.units[unitName] = &u

	return &u
}

// AddSignUnit adds signer unit to units map
func (q *Queue) AddSignUnit(unitName string, signData signer.SignData) {
	u := q.addUnit(unitName)
	// set sign data if provided
	u.signData = signData
	u.isSigningUnit = true
}

// AddVerifyUnit adds verify unit to units map
func (q *Queue) AddVerifyUnit() {
	q.addUnit(VerificationUnitName)
}

// addJob adds job to the jobs map
func (q *Queue) addJob() *Job {
	// generate unique id
	id := xid.New().String()

	// create job
	j := Job{
		ID:       id,
		TasksMap: make(map[string]Task, 1),
	}

	// add job to the jobs map
	q.jobs[id] = &j

	return &j
}

// AddSignJob adds sign job to the jobs map
func (q *Queue) AddSignJob(signConfig JobSignConfig) string {
	j := q.addJob()
	j.SignConfig = signConfig
	return j.ID
}

// AddVerifyJob adds sign job to the jobs map
func (q *Queue) AddVerifyJob() string {
	j := q.addJob()
	return j.ID
}

// DeleteJob deletes job from the jobs and database
func (q *Queue) DeleteJob(jobID string) error {
	err := q.DeleteFromDB(jobID)
	if err != nil {
		return err
	}

	delete(q.jobs, jobID)

	return nil
}

// AddTask adds task to the specific job by job id
func (q *Queue) AddTask(unitName, jobID, originalFileName, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	// check if the unit is in the map
	if _, exists := q.units[unitName]; !exists {
		return "", errors.New("unit is not in map")
	}
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job doesn't exists")
	}

	task, err := q.addTask(unitName, jobID, originalFileName, inputFilePath, outputFilePath, priority)
	if err != nil {
		return "", err
	}

	return task.ID, nil
}

func (q *Queue) addTask(unitName, jobID, originalFileName, inputFilePath, outputFilePath string, priority priority_queue.Priority) (Task, error) {
	// generate unique task id
	id := xid.New().String()

	//create task
	t := Task{
		ID:               id,
		InputFilePath:    inputFilePath,
		OutputFilePath:   outputFilePath,
		JobID:            jobID,
		Status:           StatusPending,
		OriginalFileName: originalFileName,
	}

	//create queue item
	i := priority_queue.Item{
		Value:    t,
		Priority: priority,
	}

	//add task to tasks map
	q.jobs[jobID].TasksMap[t.ID] = t

	//add item to queue
	q.units[unitName].pq.Push(i)

	return t, nil
}

func (q *Queue) AddBatchPersistentTasks(unitName, jobID string, fileNames map[string]string, priority priority_queue.Priority) error {
	// check if the unit is in the map
	if _, exists := q.units[unitName]; !exists {
		return errors.New("unit is not in map")
	}
	// check if the job is in the map
	_, exists := q.jobs[jobID]
	if !exists {
		return errors.New("job doesn't exists")
	}

	for tempFileName, originalFileName := range fileNames {
		_, err := q.addTask(unitName, jobID, originalFileName, tempFileName, tempFileName+"_signed", priority)
		if err != nil {
			return err
		}
	}

	err := q.SaveToDB(jobID)
	if err != nil {
		return err
	}

	return nil
}

// processNextTask signs task available for signing
func (q *Queue) processNextTask(unitName string) error {
	// check if the unit exists
	unit, exists := q.units[unitName]
	if !exists {
		return errors.New("signer is not in map")
	}

	// get queue
	queue := q.units[unitName]
	// get item
	item := queue.pq.Pop()
	task := item.Value.(Task)

	// get job
	job, exists := q.jobs[task.JobID]
	if !exists {
		return errors.New("signer is not in map")
	}

	// process verify or sign task
	var err error
	var verifyResp *verify.Response
	if unit.isSigningUnit {
		// sign task
		err = signTask(task, job.SignConfig, unit.signData)
	} else {
		// verify task
		verifyResp, err = verifyTask(task)
		task.VerificationData = verifyResp
	}

	// process error
	if err != nil {
		// set status
		task.Status = StatusFailed
		task.Error = err.Error()
	} else {
		// set status
		task.Status = StatusCompleted
	}

	// update tasks map
	job.TasksMap[task.ID] = task

	// increment total processed tasks
	atomic.AddUint32(&job.TotalProcesedTasks, 1)

	if len(job.TasksMap) == int(job.TotalProcesedTasks) {
		err := q.SaveToDB(job.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Queue) GetJobByID(jobID string) (Job, error) {
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job doesn't exists")
	}

	return *q.jobs[jobID], nil
}

// GetCompletedTask returns the file path if the task is completed
func (q *Queue) GetCompletedTask(jobID, taskID string) (Task, error) {
	var task Task

	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return task, errors.New("job doesn't exists")
	}

	// get task
	task, ok := q.jobs[jobID].TasksMap[taskID]
	if !ok {
		return task, errors.New("task is not found")
	}

	// check if task is not processed
	if task.Status == StatusPending {
		return task, errors.New("task is not processed yet")
	}

	// check if the stask is failed
	if task.Status == StatusFailed {
		return task, errors.New(fmt.Sprintf("task failed with error %v", task.Error))
	}

	return task, nil
}

// GetQueueSizeByUnitName returns lengths of the channels of all the priorities for the specific signer.
func (q *Queue) GetQueueSizeByUnitName(signerName string) (priority_queue.LenAll, error) {
	// check if the signer is in the map
	if _, exists := q.units[signerName]; !exists {
		return priority_queue.LenAll{}, errors.New("signer is not in map")
	}

	return q.units[signerName].pq.LenAll(), nil
}

// signTask merges job and signer signdata
func signTask(task Task, jobSignConfig JobSignConfig, signerSignData signer.SignData) error {
	// get signer sign data
	signData := signer.SignData(signerSignData)

	// merge request sign data and signer sign data
	switch {
	case jobSignConfig.Name != "":
		signData.Signature.Info.Name = jobSignConfig.Name
	case jobSignConfig.Location != "":
		signData.Signature.Info.Location = jobSignConfig.Location
	case jobSignConfig.Reason != "":
		signData.Signature.Info.Reason = jobSignConfig.Reason
	case jobSignConfig.ContactInfo != "":
		signData.Signature.Info.ContactInfo = jobSignConfig.ContactInfo
	case jobSignConfig.CertType != 0:
		signData.Signature.CertType = jobSignConfig.CertType
	case jobSignConfig.DocMDPPerms != 0:
		signData.Signature.DocMDPPerm = jobSignConfig.DocMDPPerms
	}

	err := signer.SignFile(task.InputFilePath, task.OutputFilePath, signData, jobSignConfig.ValidateSignature)
	if err != nil {
		log.WithFields(log.Fields{
			"inputFile":  task.InputFilePath,
			"outputFile": task.OutputFilePath,
			"signData":   signData,
		}).Warnf("Couldn't sign file: %s", err)

		return err
	}

	return nil
}

func verifyTask(task Task) (resp *verify.Response, err error) {
	inputFile, err := os.Open(task.InputFilePath)
	if err != nil {
		return resp, errors.Wrap(err, "")
	}
	defer inputFile.Close()

	resp, err = verify.File(inputFile)
	if err != nil {
		return resp, errors.Wrap(err, "verify task")
	}
	return resp, nil
}

// StartProcessor starts separate go routine for each signer which signs associated job tasks when they appear
func (q *Queue) StartProcessor() {
	// run separate go routine for each signer
	for _, s := range q.units {
		go func(name string) {
			for {
				// sign next task available for signing
				err := q.processNextTask(name)
				if err != nil {
					log.Printf("couldn't sign file: %v, %+v", name, err)
				}
			}
		}(s.name)
	}

}

const dbTaskPrefix = "task_"
const dbJobPrefix = "job_"

func (q *Queue) SaveToDB(jobID string) error {
	// check if the job is in the map
	job, exists := q.jobs[jobID]
	if !exists {
		return errors.New("job doesn't exists")
	}

	// save job
	marshaledJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	err = db.SaveByKey(dbJobPrefix+jobID, marshaledJob)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue) LoadFromDB() error {
	log.Info("Loading jobs from the db...")

	dbJobs, err := db.BatchLoad(dbJobPrefix)
	if err != nil {
		return errors.Wrap(err, "loading jobs from the db")
	}

	// load jobs and tasks
	for _, dbJob := range dbJobs {

		// load job
		var job Job
		err := json.Unmarshal(dbJob, &job)
		if err != nil {
			return errors.Wrap(err, "unmarshal job")
		}

		q.jobs[job.ID] = &job
	}

	return nil
}

func (q *Queue) DeleteFromDB(jobID string) error {
	// check if the job is in the map
	_, exists := q.jobs[jobID]
	if !exists {
		return errors.New("job doesn't exists")
	}

	// delete job
	err := db.DeleteByKey(dbJobPrefix + jobID)
	if err != nil {
		return err
	}

	return nil
}
