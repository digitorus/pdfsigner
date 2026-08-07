package signer

import (
	"fmt"
	"time"

	pdfsign "github.com/digitorus/pdfsign"
)

// VerifyResult is the outcome of verifying a signed PDF's signatures.
type VerifyResult struct {
	Valid      bool            `json:"valid"`
	Document   DocumentInfo    `json:"document"`
	Signatures []SignatureInfo `json:"signatures"`
}

// DocumentInfo describes metadata read from the PDF's Info dictionary.
type DocumentInfo struct {
	Title    string    `json:"title,omitempty"`
	Author   string    `json:"author,omitempty"`
	Subject  string    `json:"subject,omitempty"`
	Creator  string    `json:"creator,omitempty"`
	Producer string    `json:"producer,omitempty"`
	Pages    int       `json:"pages"`
	Created  time.Time `json:"created,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

// SignatureInfo describes the verification outcome of a single signature.
type SignatureInfo struct {
	SignerName   string           `json:"signer_name,omitempty"`
	Reason       string           `json:"reason,omitempty"`
	Location     string           `json:"location,omitempty"`
	Contact      string           `json:"contact,omitempty"`
	SigningTime  time.Time        `json:"signing_time,omitempty"`
	Valid        bool             `json:"valid"`
	TrustedChain bool             `json:"trusted_chain"`
	Revoked      bool             `json:"revoked"`
	Certificate  *CertificateInfo `json:"certificate,omitempty"`
	Timestamp    *TimestampInfo   `json:"timestamp,omitempty"`
	Errors       []string         `json:"errors,omitempty"`
	Warnings     []string         `json:"warnings,omitempty"`
}

// CertificateInfo is a JSON-friendly summary of the signer's certificate.
type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
}

// TimestampInfo describes an RFC 3161 timestamp embedded in a signature.
type TimestampInfo struct {
	Time      time.Time `json:"time"`
	Authority string    `json:"authority,omitempty"`
}

// VerifyFile verifies the signature(s) of the PDF at path.
func VerifyFile(path string) (*VerifyResult, error) {
	doc, err := pdfsign.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	v := doc.Verify()
	if err := v.Err(); err != nil {
		return nil, err
	}

	docInfo := v.Document()
	result := &VerifyResult{
		Valid: v.Valid(),
		Document: DocumentInfo{
			Title:    docInfo.Title,
			Author:   docInfo.Author,
			Subject:  docInfo.Subject,
			Creator:  docInfo.Creator,
			Producer: docInfo.Producer,
			Pages:    docInfo.Pages,
			Created:  docInfo.CreationDate,
			Modified: docInfo.ModDate,
		},
	}

	for _, sig := range v.Signatures() {
		result.Signatures = append(result.Signatures, newSignatureInfo(sig))
	}

	return result, nil
}

func newSignatureInfo(sig pdfsign.SignatureVerifyResult) SignatureInfo {
	info := SignatureInfo{
		SignerName:   sig.SignerName,
		Reason:       sig.Reason,
		Location:     sig.Location,
		Contact:      sig.Contact,
		SigningTime:  sig.SigningTime,
		Valid:        sig.Valid,
		TrustedChain: sig.TrustedChain,
		Revoked:      sig.Revoked,
	}

	if sig.Certificate != nil {
		info.Certificate = &CertificateInfo{
			Subject:      sig.Certificate.Subject.String(),
			Issuer:       sig.Certificate.Issuer.String(),
			SerialNumber: sig.Certificate.SerialNumber.String(),
			NotBefore:    sig.Certificate.NotBefore,
			NotAfter:     sig.Certificate.NotAfter,
		}
	}

	if sig.Timestamp != nil {
		info.Timestamp = &TimestampInfo{
			Time:      sig.Timestamp.Time,
			Authority: sig.Timestamp.Authority,
		}
	}

	for _, e := range sig.Errors {
		info.Errors = append(info.Errors, e.Error())
	}

	for _, w := range sig.Warnings {
		info.Warnings = append(info.Warnings, w.Error())
	}

	return info
}
