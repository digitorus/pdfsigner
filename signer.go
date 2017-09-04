package signer

import (
	"io"
	"io/ioutil"
)

// signJob is the internal job specification
type signJob struct {
	file    string // tmp file location
	options *Options
}

// Status contains the current signing proccess status for a specific document
type Status struct {
	Ready bool // true when the document is signed
}

// Options contains information required to schedule a document to be processed
// by one of the signers.
type Options struct {
	// must include info for sign.SignData
	Priority int // batch procedures should run with a low priority
}

// Signer exposes an transparent interface to the sign queue, all clients should
// implement this interface.
type Signer struct {
	q chan signJob
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

	// add this document to the signer queue
	s.q <- signJob{
		file:    tmpfile.Name(),
		options: options,
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
