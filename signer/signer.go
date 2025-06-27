package signer

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/revocation"
	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsign/verify"
	"github.com/digitorus/pdfsigner/license"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SignConfig is a SignConfig of the sign package, but with additional methods added.
type SignData sign.SignData

// SetPEM sets specific to PEM settings.
func (s *SignData) SetPEM(crtPath, keyPath, crtChainPath string) error {
	// Set certificate
	certificate_data, err := os.ReadFile(crtPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	certificate_data_block, _ := pem.Decode(certificate_data)
	if certificate_data_block == nil {
		return errors.New("failed to parse PEM block containing the certificate")
	}

	cert, err := x509.ParseCertificate(certificate_data_block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	s.Certificate = cert

	// Set key
	key_data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	key_data_block, _ := pem.Decode(key_data)
	if key_data_block == nil {
		return errors.New("failed to parse PEM block containing the private key")
	}

	pkey, err := x509.ParsePKCS1PrivateKey(key_data_block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	s.Signer = pkey

	err = s.SetCertificateChains(crtChainPath)
	if err != nil {
		return fmt.Errorf("failed to set certificate chains: %w", err)
	}

	if err := s.SetRevocationSettings(); err != nil {
		return fmt.Errorf("failed to set revocation settings: %w", err)
	}

	return nil
}

// SetCertificateChains sets certificate chain settings.
func (s *SignData) SetCertificateChains(crtChainPath string) error {
	var certificate_chains [][]*x509.Certificate

	if crtChainPath == "" {
		return nil
	}

	chain_data, err := os.ReadFile(crtChainPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate chain file: %w", err)
	}

	certificate_pool := x509.NewCertPool()
	certificate_pool.AppendCertsFromPEM(chain_data)

	certificate_chains, err = s.Certificate.Verify(x509.VerifyOptions{
		Intermediates: certificate_pool,
		CurrentTime:   s.Certificate.NotBefore,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return fmt.Errorf("failed to verify certificate chains: %w", err)
	}

	s.CertificateChains = certificate_chains

	return nil
}

// SetRevocationSettings sets default revocation settings.
func (s *SignData) SetRevocationSettings() error {
	s.RevocationData = revocation.InfoArchival{}
	s.RevocationFunction = sign.DefaultEmbedRevocationStatusFunction

	return nil
}

// SignFile checks the license, waits if limits are reached, if allowed signs the file.
func SignFile(input, output string, s SignData, validateSignature bool) error {
	// check the license and wait if limits are reached
	err := license.LD.Wait()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// set date
	if s.Signature.Info.Date.IsZero() {
		s.Signature.Info.Date = time.Now().Local()
	}

	// sign file
	err = signFile(input, output, s, validateSignature)
	if err != nil {
		return fmt.Errorf("failed to sign file: %w", err)
	}

	// log the result
	log.Debugf("File signed: %s", output)

	return err
}

func signFile(input string, output string, sign_data SignData, validateSignature bool) error {
	input_file, err := os.Open(input)
	if err != nil {
		return err
	}
	defer input_file.Close()

	output_file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer output_file.Close()

	finfo, err := input_file.Stat()
	if err != nil {
		return err
	}

	size := finfo.Size()

	rdr, err := pdf.NewReader(input_file, size)
	if err != nil {
		return err
	}

	err = sign.Sign(input_file, output_file, rdr, size, sign.SignData(sign_data))
	if err != nil {
		return err
	}

	if validateSignature {
		_, err = verify.File(output_file)
		if err != nil {
			return err
		}
	}

	return nil
}
