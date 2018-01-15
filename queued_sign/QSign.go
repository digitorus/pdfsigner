package queued_sign

import (
	"errors"
	"sync"

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/rs/xid"
)

type QSign struct {
	signers  map[string]QSigner
	sessions map[string]ThreadSafeSession
}

type QSigner struct {
	name     string
	pq       *priority_queue.PriorityQueue
	signData signer.SignData
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

func NewQSign() *QSign {
	return &QSign{
		signers:  make(map[string]QSigner, 1),
		sessions: make(map[string]ThreadSafeSession, 1),
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

func (q *QSign) NewSession(totalJobs int) string {
	id := xid.New().String()
	s := ThreadSafeSession{
		Session: &Session{ID: id, TotalJobs: totalJobs, CompletedJobsMapByID: make(map[string]Job, 1)},
		m:       &sync.Mutex{},
	}

	q.sessions[id] = s

	return id
}

func (q *QSign) PushJob(signerName, sessionID, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	if _, exists := q.signers[signerName]; !exists {
		return "", errors.New("signer is not in map")
	}

	if _, exists := q.sessions[sessionID]; !exists {
		return "", errors.New("session is not in map")
	}

	// generate unique job id
	id := xid.New().String()
	//create job
	j := Job{
		ID:             id,
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		SessionID:      sessionID,
	}

	//create queue item
	i := priority_queue.Item{
		Value:    j,
		Priority: priority,
	}

	//add item to queue
	q.signers[signerName].pq.Push(i)

	return id, nil
}

func (q *QSign) SignNextJob(signerName string) error {
	if _, exists := q.signers[signerName]; !exists {
		return errors.New("signer is not in map")
	}

	qSigner := q.signers[signerName]
	item := qSigner.pq.Pop()
	job := item.Value.(Job)

	//sign
	err := signer.SignFile(job.inputFilePath, job.outputFilePath, qSigner.signData)
	if err != nil {
		job.Error = err.Error()
	}

	// update session completed jobs
	q.addCompletedJob(job)

	return nil
}

func (q *QSign) GetSessionCompletedJobs(sessionID string) ([]Job, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return []Job{}, errors.New("session is not in map")
	}

	return q.sessions[sessionID].Session.CompletedJobs, nil
}

func (q *QSign) GetSessionByID(sessionID string) (Session, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return Session{}, errors.New("session is not in map")
	}

	return *q.sessions[sessionID].Session, nil
}

func (q *QSign) addCompletedJob(job Job) {
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

func (q *QSign) GetCompletedJobFilePath(sessionID, jobID string) (string, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return "", errors.New("session is not in map")
	}

	job, ok := q.sessions[sessionID].Session.CompletedJobsMapByID[jobID]
	if !ok {
		return "", errors.New("job not found in map")
	}

	return job.outputFilePath, nil
}

func (q *QSign) Runner() {
	for _, s := range q.signers {
		go func() {
			for {
				q.SignNextJob(s.name)
			}
		}()
	}
}
