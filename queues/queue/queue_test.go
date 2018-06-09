package queue

import (
	"io/ioutil"
	"testing"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/queues/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestQSignersMap(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	license.Load()

	// create sign data
	d := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
			},
			CertType: 2,
			Approval: false,
		},
	}
	d.SetPEM("../../testfiles/test.crt", "../../testfiles/test.pem", "")

	// create Queue
	qs := NewQueue()

	// add signer
	qs.AddSignUnit("simple", d)

	// create session
	jobID := qs.AddSignJob(SignData{})
	job, err := qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jobID, job.ID)

	// add job
	taskID, err := qs.AddTask(
		"simple",
		jobID,
		"testfile20.pdf",
		"../../testfiles/testfile20.pdf",
		"../../testfiles/testfile20_signed.pdf",
		priority_queue.HighPriority,
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(job.TasksMap))

	// sign job
	qs.processNextTask("simple")

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(job.TasksMap))
	assert.Equal(t, StatusCompleted, job.TasksMap[taskID].Status, job.TasksMap[taskID].Error)

	// test bad file
	// add job
	taskID, err = qs.AddTask(
		"simple",
		jobID,
		"malformed.pdf",
		"../../testfiles/malformed.pdf",
		"../../testfiles/malformed_signed.pdf",
		priority_queue.HighPriority,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(job.TasksMap))
	// sign job
	qs.processNextTask("simple")

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(job.TasksMap))
	assert.Equal(t, StatusFailed, job.TasksMap[taskID].Status)

	// test saving to db
	assert.NoError(t, qs.SaveToDB(jobID))

	// test load
	qs = NewQueue()
	assert.NoError(t, qs.LoadFromDB())
	jobFromDB, err := qs.GetJobByID(jobID)
	assert.NoError(t, err)
	assert.Equal(t, job, jobFromDB)

	// test delete
	assert.NoError(t, qs.DeleteJob(jobID))
	qs = NewQueue()
	assert.NoError(t, qs.LoadFromDB())
	jobFromDB, err = qs.GetJobByID(jobID)
	assert.Error(t, err)
}
