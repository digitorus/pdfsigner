package signer

import (
	"crypto"
	"io"
	"io/ioutil"
)

// Priority of signing request
type Priority int

const (
	UnknownPriority Priority = iota
	LowPriority
	MediumPriority
	HighPriority
)

// signJob is the internal job specification
type signJob struct {
	file    string // tmp file location
	options *Options
}

type priorityQueue struct {
	low    chan signJob
	medium chan signJob
	heigh  chan signJob
}

// Status contains the current signing proccess status for a specific document
type Status struct {
	Ready bool // true when the document is signed
}

// Options contains information required to schedule a document to be processed
// by one of the signers.
type Options struct {
	// must include info for sign.SignData
	Priority Priority // batch procedures should run with a low priority
}

// Signer exposes an transparent interface to the sign queue, all clients should
// implement this interface.
//
// The crypto.Signer map can contain multiple singers as defined in the config,
// a Singer implementation can be a private key or PKCS#11 device.
type Signer struct {
	q map[string]priorityQueue
	c map[string]crypto.Signer
}

// Sign reads a file and stores it at temporary location so that it can be
// processed later without consuming memeory. The function returns a tracking
// id or error.
func (s *Signer) Sign(file io.Reader, options *Options) (*string, error) {
	// TODO: Should we encrypt temporary files?
	tmpfile, err := ioutil.TempFile("", "pdf")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpfile, file)
	if err != nil {
		return nil, err
	}

	// based on the request we must identify what device or private key we should
	// be using.
	signer := "s.c-id"

	job := signJob{
		file:    tmpfile.Name(),
		options: options,
	}

	// Add job to the signing queue according to it's priority
	switch options.Priority {
	case HighPriority:
		s.q[signer].heigh <- job
	case MediumPriority:
		s.q[signer].medium <- job
	default:
		s.q[signer].low <- job
	}

	// create a unqiue id that can be used by a client to obtain the document or
	// current state of the job
	tracker := tmpfile.Name()

	return &tracker, nil
}

// Get returns the signed document based on the tracker id.
func (s *Signer) Get(tracker string) (*io.Reader, error) {
	return nil, nil
}

// Status returns if the document has been signed already.
func (s *Signer) Status(tracker string) (*Status, error) {
	// TODO: calculate retry time based on queue lenght and documents per second
	return nil, nil
}
