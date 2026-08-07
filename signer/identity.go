package signer

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Identity is the private key and certificate chain used to produce a signature.
type Identity struct {
	Signer        crypto.Signer
	Certificate   *x509.Certificate
	Intermediates []*x509.Certificate
}

// NewPEMIdentity loads a signer identity from a PEM encoded certificate and an
// RSA private key. chainPath, if non-empty, points to a PEM bundle of
// candidate intermediate/root certificates used to resolve the chain for
// Certificate.
func NewPEMIdentity(certPath, keyPath, chainPath string) (*Identity, error) {
	cert, err := readPEMCertificate(certPath)
	if err != nil {
		return nil, fmt.Errorf("read certificate: %w", err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return nil, fmt.Errorf("parse private key: failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	intermediates, err := resolveIntermediates(cert, chainPath)
	if err != nil {
		return nil, err
	}

	return &Identity{Signer: key, Certificate: cert, Intermediates: intermediates}, nil
}

func readPEMCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	return x509.ParseCertificate(block.Bytes)
}

// resolveIntermediates builds the intermediate certificate chain for cert
// from a PEM bundle of candidate certificates. It returns nil if chainPath
// is empty.
func resolveIntermediates(cert *x509.Certificate, chainPath string) ([]*x509.Certificate, error) {
	if chainPath == "" {
		return nil, nil
	}

	chainData, err := os.ReadFile(chainPath)
	if err != nil {
		return nil, fmt.Errorf("read certificate chain: %w", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(chainData)

	chains, err := cert.Verify(x509.VerifyOptions{
		Intermediates: pool,
		CurrentTime:   cert.NotBefore,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return nil, fmt.Errorf("verify certificate chain: %w", err)
	}

	if len(chains) == 0 || len(chains[0]) < 2 {
		return nil, nil
	}

	return chains[0][1:], nil // drop the leaf, keep the rest of the chain
}
