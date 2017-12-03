package queued_sign

import (
	"errors"

	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/rs/xid"
)

type QSign struct {
	signers  map[string]QSigner
	sessions map[string]Session
}

type QSigner struct {
	name     string
	pq       *priority_queue.PriorityQueue
	signData signer.SignData
}

type Session struct {
	ID            string
	TotalAdded    int
	CompletedJobs []Job
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
		sessions: make(map[string]Session, 1),
	}
}

func (q QSign) NewSession() string {
	id := xid.New().String()
	s := Session{}
	q.sessions[id] = s

	return id
}

func (q QSign) AddSigner(signerName string, signData signer.SignData, queueSize int) {
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

func (q QSign) PushJob(sessionID, signerName, inputFilePath, outputFilePath string, priority priority_queue.Priority) (string, error) {
	if _, exists := q.sessions[sessionID]; !exists {
		return "", errors.New("session is not in map")
	}

	if _, exists := q.signers[signerName]; !exists {
		return "", errors.New("signer is not in map")
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

func (q QSign) SignNextJob(signerName string) error {
	if _, exists := q.signers[signerName]; exists {
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
	session := q.sessions[job.SessionID]
	session.CompletedJobs = append(session.CompletedJobs, job)

	return nil
}

func (q QSign) Runner() {
	for _, s := range q.signers {
		go func() {
			for {
				q.SignNextJob(s.name)
			}
		}()
	}
}
