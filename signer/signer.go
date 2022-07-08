package signer

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/verify"
	log "github.com/sirupsen/logrus"

	"github.com/digitorus/pdfsign/revocation"
	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pkcs11"
	"github.com/pkg/errors"
)

// SignConfig is a SignConfig of the sign package, but with additional methods added
type SignData sign.SignData

// SetPEM sets specific to PEM settings
func (s *SignData) SetPEM(crtPath, keyPath, crtChainPath string) {
	// Set certificate
	certificate_data, err := ioutil.ReadFile(crtPath)
	if err != nil {
		log.Fatal(err)
	}
	certificate_data_block, _ := pem.Decode(certificate_data)
	if certificate_data_block == nil {
		log.Fatal("failed to parse PEM block containing the certificate")
	}
	cert, err := x509.ParseCertificate(certificate_data_block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	s.Certificate = cert

	// Set key
	key_data, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}
	key_data_block, _ := pem.Decode(key_data)
	if key_data_block == nil {
		log.Fatal("failed to parse PEM block containing the private key")
	}
	pkey, err := x509.ParsePKCS1PrivateKey(key_data_block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	s.Signer = pkey

	s.SetCertificateChains(crtChainPath)
	s.SetRevocationSettings()
}

// SetPKSC11 sets specific to PKSC11 settings
func (s *SignData) SetPKSC11(libPath, pass, crtChainPath string) {
	// pkcs11 key
	lib, err := pkcs11.FindLib(libPath)
	if err != nil {
		log.Fatal(err)
	}

	// Load Library
	ctx := pkcs11.New(lib)
	if ctx == nil {
		log.Fatal("Failed to load library")
	}
	err = ctx.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	// login
	session, err := pkcs11.CreateSession(ctx, 0, pass, false)
	if err != nil {
		log.Fatal(err)
	}
	// select the first certificate
	cert, ckaId, err := pkcs11.GetCert(ctx, session, nil)
	if err != nil {
		log.Fatal(err)
	}
	s.Certificate = cert

	// private key
	pkey, err := pkcs11.InitPrivateKey(ctx, session, ckaId)
	if err != nil {
		log.Fatal(err)
	}
	s.Signer = pkey

	s.SetCertificateChains(crtChainPath)
	s.SetRevocationSettings()
}

// SetCertificateChains sets certificate chain settings
func (s *SignData) SetCertificateChains(crtChainPath string) {
	certificate_chains := make([][]*x509.Certificate, 0)
	if crtChainPath == "" {
		return
	}

	chain_data, err := ioutil.ReadFile(crtChainPath)
	if err != nil {
		log.Fatal(err)
	}
	certificate_pool := x509.NewCertPool()
	certificate_pool.AppendCertsFromPEM(chain_data)
	certificate_chains, err = s.Certificate.Verify(x509.VerifyOptions{
		Intermediates: certificate_pool,
		CurrentTime:   s.Certificate.NotBefore,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		log.Fatal(err)
	}
	s.CertificateChains = certificate_chains
}

// SetRevocationSettings sets default revocation settings
func (s *SignData) SetRevocationSettings() {
	s.RevocationData = revocation.InfoArchival{}
	s.RevocationFunction = sign.DefaultEmbedRevocationStatusFunction
}

// SignFile checks the license, waits if limits are reached, if allowed signs the file
func SignFile(input, output string, s SignData, validateSignature bool) error {
	// check the license and wait if limits are reached
	err := license.LD.Wait()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// set date
	s.Signature.Info.Date = time.Now().Local()

	// sign file
	err = signFile(input, output, s, validateSignature)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// log the result
	log.Println("File signed:", output)

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
