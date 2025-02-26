package queue

import (
	"io"
	"testing"

	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/queues/priority_queue"
	"github.com/digitorus/pdfsigner/signer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestQSignersMap(t *testing.T) {
	logrus.SetOutput(io.Discard)

	err := license.Initialize([]byte(license.TestLicense))
	if err != nil {
		t.Fatal(err)
	}

	// create sign data
	d := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
	}
	d.SetPEM("../../testfiles/test.crt", "../../testfiles/test.pem", "")

	// create Queue
	qs := NewQueue()

	// add signer
	qs.AddSignUnit("simple", d)

	// add sign job
	jobID := qs.AddSignJob(JobSignConfig{ValidateSignature: true})

	job, err := qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, jobID, job.ID)

	// add job
	taskID, err := qs.AddTask(
		"simple",
		jobID,
		"testfile12.pdf",
		"../../testfiles/testfile12.pdf",
		"../../testfiles/testfile12_signed.pdf",
		priority_queue.HighPriority,
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, job.TasksMap, 1)

	// sign job
	assert.NoError(t, qs.processNextTask("simple"))

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, job.TasksMap, 1)
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

	assert.Len(t, job.TasksMap, 2)
	// sign job
	assert.NoError(t, qs.processNextTask("simple"))

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, job.TasksMap, 2)
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
	assert.Nil(t, jobFromDB.TasksMap)
}
