package signer

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
)

func TestSigner(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	err := license.Load()
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

	for i := 0; i < 1; i++ {
		err = SignFile("../testfiles/testfile12.pdf", "../testfiles/testfile12_signed.pdf", signData, true)
		if err != nil {
			t.Fatal(err)
		}
	}
}
