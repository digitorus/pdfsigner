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
	ID            string
	TotalJobs     int
	CompletedJobs []Job
	IsCompleted   bool
}

type Job struct {
	ID             string
	SessionID      string
	inputFilePath  string
	outputFilePath string
}

func NewQSign() QSign {
	return QSign{
		signers:  make(map[string]QSigner, 1),
		sessions: make(map[string]ThreadSafeSession, 1),
	}
}

func (q *QSign) NewSession(totalJobs int) string {
	id := xid.New().String()
	s := ThreadSafeSession{
		Session: &Session{ID: id, TotalJobs: totalJobs},
		m:       &sync.Mutex{},
	}

	q.sessions[id] = s

	return id
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
		return err
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
	s.CompletedJobs = append(s.CompletedJobs, job)
	s.IsCompleted = s.TotalJobs == len(s.CompletedJobs)
	*q.sessions[job.SessionID].Session = s

	q.sessions[job.SessionID].m.Unlock()
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
