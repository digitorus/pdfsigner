//go:build cgo

package signer

import (
	"fmt"

	"github.com/digitorus/pkcs11"
)

// NewPKCS11Identity loads a signer identity from a certificate and private
// key held on a PKCS#11 token (HSM/smart card).
func NewPKCS11Identity(libPath, pin, chainPath string) (*Identity, error) {
	lib, err := pkcs11.FindLib(libPath)
	if err != nil {
		return nil, fmt.Errorf("find pkcs11 library: %w", err)
	}

	ctx := pkcs11.New(lib)
	if ctx == nil {
		return nil, fmt.Errorf("load pkcs11 library %q", libPath)
	}

	if err := ctx.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize pkcs11 module: %w", err)
	}

	session, err := pkcs11.CreateSession(ctx, 0, pin, false)
	if err != nil {
		return nil, fmt.Errorf("open pkcs11 session: %w", err)
	}

	cert, ckaID, err := pkcs11.GetCert(ctx, session, nil)
	if err != nil {
		return nil, fmt.Errorf("read pkcs11 certificate: %w", err)
	}

	key, err := pkcs11.InitPrivateKey(ctx, session, ckaID)
	if err != nil {
		return nil, fmt.Errorf("load pkcs11 private key: %w", err)
	}

	intermediates, err := resolveIntermediates(cert, chainPath)
	if err != nil {
		return nil, err
	}

	return &Identity{Signer: key, Certificate: cert, Intermediates: intermediates}, nil
}
