package queue

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	errors2 "github.com/pkg/errors"
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
}

// unit represents queue unit which could be a signer or verifier
type unit struct {
	// name represents the name of the signer
	name string
	// signData represents sign data of the signer
	signData signer.SignData
	// pq represents priority queue used by the signer
	pq *priority_queue.PriorityQueue
	// isSigningUnit should be set to true if the unit is used for signing or false for verification
	isSigningUnit bool
}

// Job represents a job for sign queue, stores tasks and sign data to override units initial sign data
type Job struct {
	// ID represents id of the job
	ID string `json:"id"`
	// TasksMap represents tasks added to the job
	TasksMap map[string]Task `json:"-"`
	// signData represents additional sign data added by request to override signer initial sign data
	signData signer.SignData `json:"-"`
	// totalAddedTasks represents total added tasks to the job, incremented atomically
	totalAddedTasks uint32
	// totalProcessedTasks represents total processed tasks of the job, incremented atomically
	totalProcesedTasks uint32

	lastTaskAddedTime time.Time
}

// Task represents a single unit of work(file)
type Task struct {
	// ID represents id of the task
	ID string `json:"id"`
	// JobID represents id of the job task is assigned to
	JobID string `json:"-"`
	// inputFilePath represents path to the unprocessed file
	inputFilePath string
	// outputFilePath represents path to the processed file
	outputFilePath string
	// Status represents the status of the task. Pending, Failed, Completed.
	Status string `json:"status"`
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
		status = StatusCompleted
	default:
		// fail if the status is not in the list
		return []Task{}, errors.New("status is not correct")
	}

	// find tasks by status
	var tasks []Task
	for _, t := range j.TasksMap {
		if t.Status == status {
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
func (q *Queue) AddVerifyUnit(unitName string) {
	q.addUnit(unitName)
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
func (q *Queue) AddSignJob(signData signer.SignData) string {
	j := q.addJob()
	j.signData = signData
	return j.ID
}

// AddVerifyJob adds sign job to the jobs map
func (q *Queue) AddVerifyJob() string {
	j := q.addJob()
	return j.ID
}

// DeleteJob deletes job from the jobs map
func (q *Queue) DeleteJob(jobID string) {
	delete(q.jobs, jobID)
}

// AddTask adds task to the specific job by job id
func (q *Queue) AddTask(unitName, jobID, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	// check if the unit is in the map
	if _, exists := q.units[unitName]; !exists {
		return "", errors.New("unit is not in map")
	}
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job is not in map")
	}

	// generate unique task id
	id := xid.New().String()

	//create task
	t := Task{
		ID:             id,
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		JobID:          jobID,
		Status:         StatusPending,
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

	// increment total added tasks to job
	atomic.AddUint32(&q.jobs[jobID].totalAddedTasks, 1)

	return id, nil
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

	job, exists := q.jobs[task.JobID]
	if !exists {
		return errors.New("signer is not in map")
	}

	// verify or sign task
	var err error
	if unit.isSigningUnit {
		err = signTask(task, job.signData, unit.signData)
	} else {
		err = verifyTask(task)
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
	atomic.AddUint32(&job.totalProcesedTasks, 1)

	if len(job.TasksMap) == int(job.totalProcesedTasks) {
		// save state to the db after one second or close db session
	}

	return nil
}

func (q *Queue) saveToDB() {

}

func (q *Queue) loadFromDB() {

}

func (q *Queue) GetJobByID(jobID string) (Job, error) {
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job is not in map")
	}

	return *q.jobs[jobID], nil
}

// GetCompletedTaskFilePath returns the file path if the task is completed
func (q *Queue) GetCompletedTaskFilePath(jobID, taskID string) (string, error) {
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job is not in map")
	}

	// get task
	task, ok := q.jobs[jobID].TasksMap[taskID]
	if !ok {
		return "", errors.New("task not found in map")
	}

	// check if task is not processed
	if task.Status == StatusPending {
		return "", errors.New("task is not processed yet")
	}

	// check if the stask is failed
	if task.Status == StatusFailed {
		return "", errors.New(fmt.Sprintf("task failed with error: %v", task.Error))
	}

	return task.outputFilePath, nil
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
func signTask(task Task, jobSignData signer.SignData, signerSignData signer.SignData) error {
	// get signer sign data
	signData := signer.SignData(signerSignData)

	// merge request sign data and signer sign data
	switch {
	case jobSignData.Signature.Info.Name != "":
		signData.Signature.Info.Name = jobSignData.Signature.Info.Name
	case jobSignData.Signature.Info.Location != "":
		signData.Signature.Info.Location = jobSignData.Signature.Info.Location
	case jobSignData.Signature.Info.Reason != "":
		signData.Signature.Info.Reason = jobSignData.Signature.Info.Reason
	case jobSignData.Signature.Info.ContactInfo != "":
		signData.Signature.Info.ContactInfo = jobSignData.Signature.Info.ContactInfo
	case jobSignData.Signature.CertType != 0:
		signData.Signature.CertType = jobSignData.Signature.CertType
	case jobSignData.Signature.Approval != signData.Signature.Approval:
		signData.Signature.Approval = jobSignData.Signature.Approval
	}

	err := signer.SignFile(task.inputFilePath, task.outputFilePath, signData)
	if err != nil {
		log.WithFields(log.Fields{
			"inputFile":  task.inputFilePath,
			"outputFile": task.outputFilePath,
			"signData":   signData,
		}).Warnf("Couldn't sign file: %s", err)

		return err
	}

	return nil
}

func verifyTask(task Task) error {
	inputFile, err := os.Open(task.inputFilePath)
	if err != nil {
		return errors2.Wrap(err, "")
	}
	defer inputFile.Close()

	_, err = verify.Verify(inputFile)
	if err != nil {
		return err
	}
	return nil
}

// Runner starts separate go routine for each signer which signs associated job tasks when they appear
func (q *Queue) Runner() {
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
