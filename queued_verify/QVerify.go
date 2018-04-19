package queued_verify

import (
	"errors"
	"log"
	"os"
	"sync"

	"bitbucket.org/digitorus/pdfsign/verify"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"github.com/rs/xid"
)

// TODO: rewrite the same way as sign queue
// QVerify represents verification queue
type QVerify struct {
	jobs map[string]ThreadSafeJob
	pq   *priority_queue.PriorityQueue
}

type ThreadSafeJob struct {
	Job *Job
	m   *sync.Mutex
}

type Job struct {
	ID                    string          `json:"id"`
	Total                 int             `json:"total_tasks"`
	CompletedTasks        []Task          `json:"completed_task"`
	CompletedTasksMapByID map[string]Task `json:"-"`
	IsCompleted           bool            `json:"job_is_completed"`
	HadErrors             bool            `json:"had_errors"`
}

type Task struct {
	ID             string `json:"id"`
	Error          string `json:"error,omitempty"`
	JobID          string `json:"-"`
	inputFilePath  string `json:"-"`
	outputFilePath string `json:"-"`
}

func NewQVerify() *QVerify {
	return &QVerify{
		jobs: make(map[string]ThreadSafeJob, 1),
		pq:   priority_queue.New(10),
	}
}

func (q *QVerify) AddJob(totalTasks int) string {
	id := xid.New().String()
	s := ThreadSafeJob{
		Job: &Job{
			ID:    id,
			Total: totalTasks,
			CompletedTasksMapByID: make(map[string]Task, 1),
		},
		m: &sync.Mutex{},
	}

	q.jobs[id] = s

	return id
}

func (q *QVerify) AddTask(jobID, inputFilePath string, priority priority_queue.Priority) (string, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return "", errors.New("job is not in map")
	}

	// generate unique task id
	id := xid.New().String()
	//create task
	j := Task{
		ID:            id,
		inputFilePath: inputFilePath,
		JobID:         jobID,
	}

	//create queue item
	i := priority_queue.Item{
		Value:    j,
		Priority: priority,
	}

	//add item to queue
	q.pq.Push(i)

	return id, nil
}

func (q *QVerify) VerifyNextTask() error {
	item := q.pq.Pop()
	task := item.Value.(Task)

	inputFile, err := os.Open(task.inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	_, err = verify.Verify(inputFile)
	if err != nil {
		task.Error = err.Error()
	}

	// update job completed tasks
	q.addCompletedTask(task)

	return nil
}

func (q *QVerify) GetCompletedJobTasks(jobID string) ([]Task, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return []Task{}, errors.New("job is not in map")
	}

	return q.jobs[jobID].Job.CompletedTasks, nil
}

func (q *QVerify) GetJobByID(jobID string) (Job, error) {
	if _, exists := q.jobs[jobID]; !exists {
		return Job{}, errors.New("job is not in map")
	}

	return *q.jobs[jobID].Job, nil
}

func (q *QVerify) addCompletedTask(task Task) {
	q.jobs[task.JobID].m.Lock()

	s := *q.jobs[task.JobID].Job

	// update values
	s.CompletedTasks = append(s.CompletedTasks, task)
	s.CompletedTasksMapByID[task.ID] = task
	s.IsCompleted = s.Total == len(s.CompletedTasks)
	if task.Error != "" {
		s.HadErrors = true
	}

	// update job
	*q.jobs[task.JobID].Job = s

	q.jobs[task.JobID].m.Unlock()
}

func (q *QVerify) Runner() {
	go func() {
		for {
			q.VerifyNextTask()
		}
	}()
}
