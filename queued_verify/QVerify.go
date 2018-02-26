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

type QVerify struct {
	sessions map[string]ThreadSafeSession
	pq       *priority_queue.PriorityQueue
}

type ThreadSafeSession struct {
	Session *Session
	m       *sync.Mutex
}

type Session struct {
	ID                   string         `json:"id"`
	TotalJobs            int            `json:"total_jobs"`
	CompletedJobs        []Job          `json:"completed_jobs"`
	CompletedJobsMapByID map[string]Job `json:"-"`
	IsCompleted          bool           `json:"session_is_completed"`
	HadErrors            bool           `json:"had_errors"`
}

type Job struct {
	ID             string `json:"id"`
	Error          string `json:"error,omitempty"`
	SessionID      string `json:"-"`
	inputFilePath  string `json:"-"`
	outputFilePath string `json:"-"`
}

func NewQVerify() *QVerify {
	return &QVerify{
		sessions: make(map[string]ThreadSafeSession, 1),
		pq:       priority_queue.New(10),
	}
}

func (q *QVerify) NewSession(totalJobs int) string {
	id := xid.New().String()
	s := ThreadSafeSession{
		Session: &Session{
			ID:                   id,
			TotalJobs:            totalJobs,
			CompletedJobsMapByID: make(map[string]Job, 1),
		},
		m: &sync.Mutex{},
	}

	q.sessions[id] = s

	return id
}

func (q *QVerify) PushJob(sessionID, inputFilePath string, priority priority_queue.Priority) (string, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return "", errors.New("session is not in map")
	}

	// generate unique job id
	id := xid.New().String()
	//create job
	j := Job{
		ID:            id,
		inputFilePath: inputFilePath,
		SessionID:     sessionID,
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

func (q *QVerify) VerifyNextJob() error {
	item := q.pq.Pop()
	job := item.Value.(Job)

	inputFile, err := os.Open(job.inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	_, err = verify.Verify(inputFile)
	if err != nil {
		job.Error = err.Error()
	}

	// update session completed jobs
	q.addCompletedJob(job)

	return nil
}

func (q *QVerify) GetSessionCompletedJobs(sessionID string) ([]Job, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return []Job{}, errors.New("session is not in map")
	}

	return q.sessions[sessionID].Session.CompletedJobs, nil
}

func (q *QVerify) GetSessionByID(sessionID string) (Session, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return Session{}, errors.New("session is not in map")
	}

	return *q.sessions[sessionID].Session, nil
}

func (q *QVerify) addCompletedJob(job Job) {
	q.sessions[job.SessionID].m.Lock()

	s := *q.sessions[job.SessionID].Session

	// update values
	s.CompletedJobs = append(s.CompletedJobs, job)
	s.CompletedJobsMapByID[job.ID] = job
	s.IsCompleted = s.TotalJobs == len(s.CompletedJobs)
	if job.Error != "" {
		s.HadErrors = true
	}

	// update session
	*q.sessions[job.SessionID].Session = s

	q.sessions[job.SessionID].m.Unlock()
}

func (q *QVerify) Runner() {
	go func() {
		for {
			q.VerifyNextJob()
		}
	}()
}
