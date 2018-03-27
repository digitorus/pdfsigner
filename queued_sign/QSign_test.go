package queued_sign

import (
	"testing"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/stretchr/testify/assert"
)

func TestQSignersMap(t *testing.T) {
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
	d.SetPEM("../testfiles/test.crt", "../testfiles/test.pem", "")

	// create QSign
	qs := NewQSign()

	// add signer
	qs.AddSigner("simple", d, 10)

	// create session
	var signData signer.SignData
	jobID := qs.AddJob(1, signData)
	job, err := qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, 1, job.TotalTasks)

	// add job
	qs.AddTask(
		"simple",
		jobID,
		"../testfiles/testfile20.pdf",
		"../testfiles/testfile20_signed.pdf",
		priority_queue.HighPriority,
	)

	// sign job
	qs.SignNextTask("simple")

	job, err = qs.GetJobByID(jobID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, job.IsCompleted)
	assert.Equal(t, 1, len(job.CompletedTasks))
	assert.Equal(t, "", job.CompletedTasks[0].Error)
}
