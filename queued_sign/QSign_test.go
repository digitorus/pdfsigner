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
	sessionID := qs.NewSession(1)
	session, err := qs.GetSessionByID(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, 1, session.TotalJobs)

	// add job
	qs.PushJob(
		"simple",
		sessionID,
		"../testfiles/testfile20.pdf",
		"../testfiles/testfile20_signed.pdf",
		priority_queue.HighPriority,
	)

	// sign job
	qs.SignNextJob("simple")

	session, err = qs.GetSessionByID(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, session.IsCompleted)
	assert.Equal(t, 1, len(session.CompletedJobs))
	assert.Equal(t, "", session.CompletedJobs[0].Error)
}
