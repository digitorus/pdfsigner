package signer

import (
	"log"
	"testing"

	"bitbucket.org/digitorus/pdfsign/sign"
)

func TestSigner(t *testing.T) {
	d := SignData{
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

	err := SignFile("../testfiles/testfile20.pdf", "../testfiles/testfile20_signed.pdf", d)
	if err != nil {
		log.Fatal(err)
	}
}
