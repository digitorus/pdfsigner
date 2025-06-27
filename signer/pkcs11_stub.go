//go:build !cgo
// +build !cgo

package signer

import "github.com/pkg/errors"

// SetPKSC11 provides a stub implementation when PKCS11 is not available.
func (s *SignData) SetPKSC11(libPath, pass, crtChainPath string) error {
	return errors.New("PKCS11 support is not available in this build")
}
