//go:build cgo
// +build cgo

package signer

import (
	"errors"
	"fmt"

	"github.com/digitorus/pkcs11"
)

// SetPKSC11 sets specific to PKSC11 settings.
func (s *SignData) SetPKSC11(libPath, pass, crtChainPath string) error {
	// pkcs11 key
	lib, err := pkcs11.FindLib(libPath)
	if err != nil {
		return fmt.Errorf("failed to find PKCS11 library: %w", err)
	}

	// Load Library
	ctx := pkcs11.New(lib)
	if ctx == nil {
		return errors.New("failed to load PKCS11 library")
	}

	err = ctx.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize PKCS11 context: %w", err)
	}

	// login
	session, err := pkcs11.CreateSession(ctx, 0, pass, false)
	if err != nil {
		return fmt.Errorf("failed to create PKCS11 session: %w", err)
	}

	// select the first certificate
	cert, ckaId, err := pkcs11.GetCert(ctx, session, nil)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}

	s.Certificate = cert

	// private key
	pkey, err := pkcs11.InitPrivateKey(ctx, session, ckaId)
	if err != nil {
		return fmt.Errorf("failed to initialize private key: %w", err)
	}

	s.Signer = pkey

	if err := s.SetCertificateChains(crtChainPath); err != nil {
		return fmt.Errorf("failed to set certificate chains: %w", err)
	}

	if err := s.SetRevocationSettings(); err != nil {
		return fmt.Errorf("failed to set revocation settings: %w", err)
	}

	return nil
}
