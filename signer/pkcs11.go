//go:build cgo
// +build cgo

package signer

import (
	"github.com/digitorus/pkcs11"
	log "github.com/sirupsen/logrus"
)

// SetPKSC11 sets specific to PKSC11 settings.
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
