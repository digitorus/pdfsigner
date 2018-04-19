package queuedsign

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

// QSign represents sign queue
type QSign struct {
	// signers represent all the signers by name of the signer
	signers map[string]QSigner
	// jobs represents jobs by id of the job
	jobs map[string]*Job
}

// QSigner represents
type QSigner struct {
	name     string
	pq       *priority_queue.PriorityQueue
	signData signer.SignData
}

type Job struct {
	// ID represents id of the job
	ID string `json:"id"`
	// TotalTasks represents total number of tasks to add
	TotalTasks int             `json:"total"`
	TasksMap   map[string]Task `json:"-"`
	SignData   signer.SignData `json:"-"`
}

func (j *Job) GetTasks(status string) ([]Task, error) {
	switch status {
	case StatusCompleted:
	case StatusFailed:
	case StatusPending:
	case "":
		status = StatusCompleted
	default:
		return []Task{}, errors.New("status is not correct")
	}

	var tasks []Task
	for _, t := range j.TasksMap {
		if t.Status == status {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

type Task struct {
	ID             string `json:"id"`
	JobID          string `json:"-"`
	inputFilePath  string
	outputFilePath string
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
}

func NewQSign() *QSign {
	return &QSign{
		signers: make(map[string]QSigner, 1),
		jobs:    make(map[string]*Job, 1),
	}
}

// AddSigner adds signer to the queue signers pool
func (q *QSign) AddSigner(signerName string, signData signer.SignData, queueSize int) {
	// skip if already setup
	if _, exists := q.signers[signerName]; exists {
		return
	}

	qs := QSigner{
		name:     signerName,
		pq:       priority_queue.New(queueSize),
		signData: signData,
	}

	q.signers[signerName] = qs
}

func (q *QSign) AddJob(totalTasks int, signData signer.SignData) string {
	id := xid.New().String()
	j := &Job{
		ID:         id,
		TotalTasks: totalTasks,
		SignData:   signData,
		TasksMap:   make(map[string]Task, 1),
	}

	q.jobs[id] = j

	return id
}

func (q *QSign) DeleteJob(jobID string) {
	delete(q.jobs, jobID)
}

func (q *QSign) AddTask(signerName, jobID, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	if _, exists := q.signers[signerName]; !exists {
		return "", errors.New("signer is not in map")
	}

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

func (q *QSign) SignNextTask(signerName string) error {
	if _, exists := q.signers[signerName]; !exists {
		return errors.New("signer is not in map")
	}

	qSigner := q.signers[signerName]
	item := qSigner.pq.Pop()
	task := item.Value.(Task)

	signData, err := q.mergeJobSignerSignData(task.JobID, qSigner.signData)
	if err != nil {
		return err
	}

	//sign
	err = signer.SignFile(task.inputFilePath, task.outputFilePath, signData)
	if err != nil {
		log.Printf("Couldn't sign file: %v, %+v", task.inputFilePath, err)
		task.Status = StatusFailed
		task.Error = err.Error()
	} else {
		task.Status = StatusCompleted
	}

	// update tasks map
	q.jobs[task.JobID].TasksMap[task.ID] = task

	return nil
}

func (q *QSign) mergeJobSignerSignData(jobID string, signerSignData signer.SignData) (signer.SignData, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return signerSignData, errors.New("job is not in map")
	}

	jobSignData := q.jobs[jobID].SignData
	signData := signer.SignData(signerSignData)

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

func (q *QSign) GetJobByID(jobID string) (Job, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job is not in map")
	}

	return *q.jobs[jobID], nil
}

func (q *QSign) GetCompletedTaskFilePath(jobID, taskID string) (string, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job is not in map")
	}

	task, ok := q.jobs[jobID].TasksMap[taskID]
	if !ok {
		return "", errors.New("task not found in map")
	}

	if task.Status == StatusPending {
		return "", errors.New("task is not processed yet")
	}

	if task.Status == StatusFailed {
		return "", errors.New(fmt.Sprintf("task failed with error: %v", task.Error))
	}

	return task.outputFilePath, nil
}

func (q *QSign) GetQueueSizeBySignerName(signerName string) (priority_queue.LenAll, error) {
	if _, exists := q.signers[signerName]; !exists {
		return priority_queue.LenAll{}, errors.New("signer is not in map")
	}

	return q.signers[signerName].pq.LenAll(), nil
}

func (q *QSign) Runner() {
	for _, s := range q.signers {
		go func(name string) {
			for {
				err := q.SignNextTask(name)
				if err != nil {
					log.Printf("couldn't sign file: %v, %+v", name, err)
				}
			}
		}(s.name)
	}
}
