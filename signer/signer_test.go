package signer

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
)

func TestSigner(t *testing.T) {
	// create signer
	signData := SignData{
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
	signData.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")

	err := SignFile("../testfiles/testfile20.pdf", "../testfiles/testfile20_signed.pdf", signData)
	if err != nil {
		t.Fatal(err)
	}

	err = SignFile("../testfiles/testfile20.pdf", "../testfiles/testfile20_signed.pdf", signData)
	if err != nil {
		t.Fatal(err)
	}
}
