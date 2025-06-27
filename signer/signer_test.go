package signer

import (
	"testing"

	"github.com/digitorus/pdfsign/sign"
)

func TestSigner(t *testing.T) {
	// create signer
	signData := SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "John Doe",
				Location:    "Amsterdam, NL",
				Reason:      "Document Approval",
				ContactInfo: "john.doe@example.com",
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
	}

	err := signData.SetPEM("../testfiles/test.crt", "../testfiles/test.pem", "")
	if err != nil {
		t.Fatal(err)
	}

	for range 1 {
		err = SignFile("../testfiles/testfile12.pdf", "../testfiles/testfile12_signed.pdf", signData, true)
		if err != nil {
			t.Fatal(err)
		}
	}
}
