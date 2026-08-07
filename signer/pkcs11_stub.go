//go:build !cgo

package signer

import "errors"

// NewPKCS11Identity is a stub used when PKCS11 support is not available.
func NewPKCS11Identity(libPath, pin, chainPath string) (*Identity, error) {
	return nil, errors.New("PKCS11 support is not available in this build; rebuild with CGO_ENABLED=1")
}
