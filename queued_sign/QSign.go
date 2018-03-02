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
	ID                   string          `json:"id"`
	TotalJobs            int             `json:"total_jobs"`
	CompletedJobs        []Job           `json:"completed_jobs"`
	CompletedJobsMapByID map[string]Job  `json:"-"`
	IsCompleted          bool            `json:"session_is_completed"`
	HadErrors            bool            `json:"had_errors"`
	SignData             signer.SignData `json:"-"`
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

func (q *QSign) NewSession(totalJobs int, signData signer.SignData) string {
	id := xid.New().String()
	s := ThreadSafeSession{
		Session: &Session{
			ID:                   id,
			TotalJobs:            totalJobs,
			SignData:             signData,
			CompletedJobsMapByID: make(map[string]Job, 1),
		},
		m: &sync.Mutex{},
	}

	q.sessions[id] = s

	return id
}

func (q *QSign) DeleteSession(sessionID string) {
	delete(q.sessions, sessionID)
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

	signData, err := q.mergeSessionSignerSignData(job.SessionID, qSigner.signData)
	if err != nil {
		return err
	}

	//sign
	err = signer.SignFile(job.inputFilePath, job.outputFilePath, signData)
	if err != nil {
		log.Printf("Couldn't sign file: %v, %+v", job.inputFilePath, err)
		job.Error = err.Error()
	} else {
		log.Println("File signed:", job.outputFilePath)
	}

	// update session completed jobs
	q.addCompletedJob(job)

	return nil
}

func (q *QSign) mergeSessionSignerSignData(sessionID string, signerSignData signer.SignData) (signer.SignData, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return signerSignData, errors.New("session is not in map")
	}

	sessionSignData := q.sessions[sessionID].Session.SignData
	signData := signer.SignData(signerSignData)

	switch {
	case sessionSignData.Signature.Info.Name != "":
		signData.Signature.Info.Name = sessionSignData.Signature.Info.Name
	case sessionSignData.Signature.Info.Location != "":
		signData.Signature.Info.Location = sessionSignData.Signature.Info.Location
	case sessionSignData.Signature.Info.Reason != "":
		signData.Signature.Info.Reason = sessionSignData.Signature.Info.Reason
	case sessionSignData.Signature.Info.ContactInfo != "":
		signData.Signature.Info.ContactInfo = sessionSignData.Signature.Info.ContactInfo
	case sessionSignData.Signature.CertType != 0:
		signData.Signature.CertType = sessionSignData.Signature.CertType
	case sessionSignData.Signature.Approval != signData.Signature.Approval:
		signData.Signature.Approval = sessionSignData.Signature.Approval
	}

	return signData, nil
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
				err := q.SignNextJob(s.name)
				if err != nil {
					log.Printf("Couldn't sign file: %v, %+v", s.name, err)
				}
			}
		}()
	}
}
