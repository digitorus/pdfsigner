//go:build !cgo
// +build !cgo

package signer

import (
	log "github.com/sirupsen/logrus"
)

// SetPKSC11 provides a stub implementation when PKCS11 is not available.
func (s *SignData) SetPKSC11(libPath, pass, crtChainPath string) {
	log.Fatal("PKCS11 support is not available in this build. Please rebuild with CGO_ENABLED=1")
}
