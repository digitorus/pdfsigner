package queued_sign

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
)

func TestQSignersMap(t *testing.T) {
	d := signer.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType: 2,
			Approval: false,
		},
	}

	d.SetPEM("../testfiles/test.crt", "../testfiles/test.pem", "")

	qs := NewQSign()
	qs.AddSigner("simple", d, 10)
	sessionID := qs.NewSession()
	qs.PushJob(
		sessionID,
		"simple",
		"../testfiles/testfile20.pdf",
		"../testfiles/testfile20_signed.pdf",
		priority_queue.HighPriority,
	)

	qs.SignNextJob("simple")
}
