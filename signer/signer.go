package signer

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"bitbucket.org/digitorus/pdfsign/revocation"
	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pkcs11"
	errors2 "github.com/pkg/errors"
)

type SignData sign.SignData

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

func (s *SignData) SetRevocationSettings() {
	s.RevocationData = revocation.InfoArchival{}
	s.RevocationFunction = sign.DefaultEmbedRevocationStatusFunction
}

func SignFile(input, output string, s SignData) error {
	if time.Now().After(license.LD.End) {
		return errors2.Wrap(errors.New(fmt.Sprintf("license is valid until:%v, please update the license", license.LD.End)), "")
	}

	for {
		if license.LD.RL.Allow() {
			break
		} else {
			left, limit := license.LD.RL.Left()
			if license.IsTotalLimit(limit) {
				return errors2.Wrap(errors.New("total license limits exceeded, please update the license"), "")
			}

			log.Println(license.ErrOverLimit, "wait for:", left)
			time.Sleep(left)
		}
	}

	s.Signature.Info.Date = time.Now().Local()
	err := sign.SignFile(input, output, sign.SignData(s))
	if err != nil {
		return errors2.Wrap(err, "")
	}

	log.Println("File signed:", output)

	return err
}
