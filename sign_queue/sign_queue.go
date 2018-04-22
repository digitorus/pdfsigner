package signqueue

import (
	"errors"
	"fmt"
	"log"

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
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
)

// SignQueue represents sign queue
type SignQueue struct {
	// signers represent all the signers by name of the signer
	signers map[string]queueSigner
	// jobs represents jobs by id of the job
	jobs map[string]*Job
}

// queueSigner represents queue signer with it's priority queue
type queueSigner struct {
	// name represents the name of the signer
	name string
	// signData represents sign data of the signer
	signData signer.SignData
	// pq represents priority queue used by the signer
	pq *priority_queue.PriorityQueue
}

// Job represents a job for sign queue, stores tasks and sign data to override signers initial sign data
type Job struct {
	// ID represents id of the job
	ID string `json:"id"`
	// TasksMap represents tasks added to the job
	TasksMap map[string]Task `json:"-"`
	// SignData represents additional sign data added by request to override signer initial sign data
	SignData signer.SignData `json:"-"`
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

// Task represents a single unit of work(file)
type Task struct {
	// ID represents id of the task
	ID string `json:"id"`
	// JobID represents id of the job task is assigned to
	JobID string `json:"-"`
	// inputFilePath represents path to the unsigned file
	inputFilePath string
	// outputFilePath represents path to the signed file
	outputFilePath string
	// Status represents the status of the task. Pending, Failed, Completed.
	Status string `json:"status"`
	// Error represents error if the task failed
	Error string `json:"error,omitempty"`
}

// NewSignQueue creates new sign queue
func NewSignQueue() *SignQueue {
	return &SignQueue{
		signers: make(map[string]queueSigner, 1),
		jobs:    make(map[string]*Job, 1),
	}
}

// AddSigner adds signer to signers map
func (q *SignQueue) AddSigner(signerName string, signData signer.SignData, queueSize int) {
	// skip if already setup
	if _, exists := q.signers[signerName]; exists {
		return
	}

	qs := queueSigner{
		name:     signerName,
		pq:       priority_queue.New(queueSize),
		signData: signData,
	}

	q.signers[signerName] = qs
}

// AddJob adds job to the jobs map
func (q *SignQueue) AddJob(signData signer.SignData) string {
	// generate unique id
	id := xid.New().String()

	// create job
	j := Job{
		ID:       id,
		SignData: signData,
		TasksMap: make(map[string]Task, 1),
	}

	// assign job to sign map
	q.jobs[id] = &j

	return id
}

// DeleteJob deletes job from the jobs map
func (q *SignQueue) DeleteJob(jobID string) {
	delete(q.jobs, jobID)
}

// AddTask adds task to the specific job by job id
func (q *SignQueue) AddTask(signerName, jobID, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	// check if the signer is in the map
	if _, exists := q.signers[signerName]; !exists {
		return "", errors.New("signer is not in map")
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
	q.signers[signerName].pq.Push(i)

	return id, nil
}

// SignNextTask signs task available for signing
func (q *SignQueue) SignNextTask(signerName string) error {
	if _, exists := q.signers[signerName]; !exists {
		return errors.New("signer is not in map")
	}

	// get queue
	qSigner := q.signers[signerName]

	// get item
	item := qSigner.pq.Pop()
	task := item.Value.(Task)

	// merge signer and request sign data
	signData, err := q.mergeJobSignerSignData(task.JobID, qSigner.signData)
	if err != nil {
		return err
	}

	// sign file
	err = signer.SignFile(task.inputFilePath, task.outputFilePath, signData)
	if err != nil {
		// log error
		log.Printf("Couldn't sign file: %v, %+v", task.inputFilePath, err)
		// set status
		task.Status = StatusFailed
		task.Error = err.Error()
	} else {
		// set status
		task.Status = StatusCompleted
	}

	// update tasks map
	q.jobs[task.JobID].TasksMap[task.ID] = task

	return nil
}

// mergeJobSignerSignData merges job and signer signdata
func (q *SignQueue) mergeJobSignerSignData(jobID string, signerSignData signer.SignData) (signer.SignData, error) {
	// check if the job is in the jobs map
	if _, exists := q.jobs[jobID]; !exists {
		return signerSignData, errors.New("job is not in map")
	}

	// get request sign data
	jobSignData := q.jobs[jobID].SignData

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

	return signData, nil
}

func (q *SignQueue) GetJobByID(jobID string) (Job, error) {
	// check if the job is in the map
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job is not in map")
	}

	return *q.jobs[jobID], nil
}

// GetCompletedTaskFilePath returns the file path if the task is completed
func (q *SignQueue) GetCompletedTaskFilePath(jobID, taskID string) (string, error) {
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

// GetQueueSizeBySignerName returns lengths of the channels of all the priorities for the specific signer.
func (q *SignQueue) GetQueueSizeBySignerName(signerName string) (priority_queue.LenAll, error) {
	// check if the signer is in the map
	if _, exists := q.signers[signerName]; !exists {
		return priority_queue.LenAll{}, errors.New("signer is not in map")
	}

	return q.signers[signerName].pq.LenAll(), nil
}

// Runner starts separate go routine for each signer which signs associated job tasks when they appear
func (q *SignQueue) Runner() {
	// run separate go routine for each signer
	for _, s := range q.signers {
		go func(name string) {
			for {
				// sign next task available for signing
				err := q.SignNextTask(name)
				if err != nil {
					log.Printf("couldn't sign file: %v, %+v", name, err)
				}
			}
		}(s.name)
	}
}
