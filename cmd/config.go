package cmd

import (
	"fmt"

	"github.com/digitorus/pdfsigner/signer"
)

// Config holds the root configuration structure.
type Config struct {
	LicensePath string                   `mapstructure:"licensePath"`
	Services    map[string]serviceConfig `mapstructure:"services"`
	Signers     map[string]signerConfig  `mapstructure:"signers"`
}

// serviceConfig is a config of the service.
type serviceConfig struct {
	Type              string   `mapstructure:"type"`
	Signer            string   `mapstructure:"signer,omitempty"`
	Signers           []string `mapstructure:"signers,omitempty"`
	In                string   `mapstructure:"in,omitempty"`
	Out               string   `mapstructure:"out,omitempty"`
	ValidateSignature bool     `mapstructure:"validateSignature"`
	Addr              string   `mapstructure:"addr,omitempty"`
	Port              string   `mapstructure:"port,omitempty"`
}

type signerConfig struct {
	Type     string          `mapstructure:"type"`
	Cert     string          `mapstructure:"cert,omitempty"`
	Key      string          `mapstructure:"key,omitempty"`
	Lib      string          `mapstructure:"lib,omitempty"`
	Pass     string          `mapstructure:"pass,omitempty"`
	Chain    string          `mapstructure:"chain,omitempty"`
	SignData signer.SignData `mapstructure:"signData"`
}

var (
	config Config
)

// getSignerConfigByName returns config of the signer by name.
func getSignerConfigByName(signer string) (signerConfig, error) {
	if signer == "" {
		return signerConfig{}, fmt.Errorf("signer name is empty")
	}

	// Try to get directly from the map first (more efficient)
	if signer, exists := config.Signers[signer]; exists {
		return signer, nil
	}

	// fail if signer not found
	return signerConfig{}, fmt.Errorf("signer not found: %s", signer)
}
