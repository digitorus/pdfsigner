package signer

import (
	"io"
	"testing"

	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/sirupsen/logrus"
)

func TestSigner(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// test initialize
	err := license.Initialize([]byte(license.TestLicense))
	if err != nil {
		t.Fatal(err)
	}

	// create signer
	signData := SignData{
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
	signData.SetPEM("../testfiles/test.crt", "../testfiles/test.pem", "")

	for range 1 {
		err = SignFile("../testfiles/testfile12.pdf", "../testfiles/testfile12_signed.pdf", signData, true)
		if err != nil {
			t.Fatal(err)
		}
	}
}
