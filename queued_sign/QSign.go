package queued_sign

import (
	"errors"
	"log"
	"sync"

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/rs/xid"
)

type QSign struct {
	signers map[string]QSigner
	jobs    map[string]ThreadSafeJob
}

type QSigner struct {
	name     string
	pq       *priority_queue.PriorityQueue
	signData signer.SignData
}

type ThreadSafeJob struct {
	Job *Job
	m   *sync.Mutex
}

type Job struct {
	ID                    string          `json:"id"`
	TotalTasks            int             `json:"total_tasks"`
	ProcessedTasks        []Task          `json:"processed_tasks"`
	ProcessedTasksMapByID map[string]Task `json:"-"`
	IsCompleted           bool            `json:"job_is_completed"`
	HadErrors             bool            `json:"had_errors"`
	SignData              signer.SignData `json:"-"`
}

type Task struct {
	ID             string `json:"id"`
	JobID          string `json:"-"`
	inputFilePath  string `json:"-"`
	outputFilePath string `json:"-"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
}

func NewQSign() *QSign {
	return &QSign{
		signers: make(map[string]QSigner, 1),
		jobs:    make(map[string]ThreadSafeJob, 1),
	}
}

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
	s := ThreadSafeJob{
		Job: &Job{
			ID:                    id,
			TotalTasks:            totalTasks,
			SignData:              signData,
			ProcessedTasksMapByID: make(map[string]Task, 1),
		},
		m: &sync.Mutex{},
	}

	q.jobs[id] = s

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
		Status:         "Pending",
	}

	//create queue item
	i := priority_queue.Item{
		Value:    t,
		Priority: priority,
	}

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
		task.Status = "Failed"
		task.Error = err.Error()
	} else {
		task.Status = "Completed"
		log.Println("File signed:", task.outputFilePath)
	}

	// update job completed tasks
	q.addProcessedTask(task)

	return nil
}

func (q *QSign) mergeJobSignerSignData(jobID string, signerSignData signer.SignData) (signer.SignData, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return signerSignData, errors.New("job is not in map")
	}

	jobSignData := q.jobs[jobID].Job.SignData
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

func (q *QSign) GetCompletedJobTasks(jobID string) ([]Task, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return []Task{}, errors.New("job is not in map")
	}

	return q.jobs[jobID].Job.ProcessedTasks, nil
}

func (q *QSign) GetJobByID(jobID string) (Job, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job is not in map")
	}

	return *q.jobs[jobID].Job, nil
}

func (q *QSign) addProcessedTask(task Task) {
	q.jobs[task.JobID].m.Lock()

	s := *q.jobs[task.JobID].Job

	// update values
	s.ProcessedTasks = append(s.ProcessedTasks, task)
	s.ProcessedTasksMapByID[task.ID] = task
	s.IsCompleted = s.TotalTasks == len(s.ProcessedTasks)
	if task.Error != "" {
		s.HadErrors = true
	}

	// update job
	*q.jobs[task.JobID].Job = s

	q.jobs[task.JobID].m.Unlock()
}

func (q *QSign) GetCompletedTaskFilePath(jobID, taskID string) (string, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job is not in map")
	}

	task, ok := q.jobs[jobID].Job.ProcessedTasksMapByID[taskID]
	if !ok {
		return "", errors.New("task not found in map")
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
		go func() {
			for {
				err := q.SignNextTask(s.name)
				if err != nil {
					log.Printf("Couldn't sign file: %v, %+v", s.name, err)
				}
			}
		}()
	}
}
