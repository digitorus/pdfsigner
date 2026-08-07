package signer

import (
	"io"
	"path/filepath"
	"testing"

	pdfsign "github.com/digitorus/pdfsign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/sirupsen/logrus"
)

func testOptions() Options {
	return Options{
		SignerName: "Tim",
		Location:   "Spain",
		Reason:     "Test",
		Contact:    "None",
		Type:       pdfsign.CertificationSignature,
		Permission: pdfsign.AllowFormFilling,
	}
}

func TestSignFile(t *testing.T) {
	logrus.SetOutput(io.Discard)

	if err := license.Initialize([]byte(license.TestLicense)); err != nil {
		t.Fatal(err)
	}

	identity, err := NewPEMIdentity("../testfiles/test.crt", "../testfiles/test.pem", "")
	if err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "testfile12_signed.pdf")

	if err := SignFile("../testfiles/testfile12.pdf", output, identity, testOptions(), true); err != nil {
		t.Fatal(err)
	}
}

func TestSignFileWithAppearance(t *testing.T) {
	logrus.SetOutput(io.Discard)

	if err := license.Initialize([]byte(license.TestLicense)); err != nil {
		t.Fatal(err)
	}

	identity, err := NewPEMIdentity("../testfiles/test.crt", "../testfiles/test.pem", "")
	if err != nil {
		t.Fatal(err)
	}

	opts := testOptions()
	// pdfsign only allows visible appearances on approval signatures.
	opts.Type = pdfsign.ApprovalSignature
	opts.Appearance = &Appearance{Page: 1, X: 20, Y: 20, Width: 200, Height: 80}

	output := filepath.Join(t.TempDir(), "testfile12_signed_visible.pdf")

	if err := SignFile("../testfiles/testfile12.pdf", output, identity, opts, true); err != nil {
		t.Fatal(err)
	}

	// The test certificate is self-signed and lacks the Digital Signature key
	// usage / Extended Key Usage extensions pdfsign's default VerifyOptions
	// require, so it never reports as trusted or fully Valid here; this only
	// asserts the signature (and its visible appearance) round-trips.
	result, err := VerifyFile(output)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(result.Signatures))
	}
}
