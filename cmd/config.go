package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/digitorus/pdfsigner/signer"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds the root configuration structure.
type Config struct {
	LicensePath string                   `mapstructure:"licensePath"`
	Services    map[string]serviceConfig `mapstructure:"services"`
	Signers     map[string]signerConfig  `mapstructure:"signers"`
}

// serviceConfig is a config of the service.
type serviceConfig struct {
	Name              string   `mapstructure:"-"` // Added for backward compatibility
	Type              string   `mapstructure:"type"`
	Signer            string   `mapstructure:"signer,omitempty"`
	Signers           []string `mapstructure:"signers,omitempty"`
	In                string   `mapstructure:"in,omitempty"`
	Out               string   `mapstructure:"out,omitempty"`
	ValidateSignature bool     `mapstructure:"validateSignature"`
	Addr              string   `mapstructure:"addr,omitempty"`
	Port              string   `mapstructure:"port,omitempty"` // Changed to string
}

type signerConfig struct {
	Name         string          `mapstructure:"-"` // Added for backward compatibility
	Type         string          `mapstructure:"type"`
	CrtPath      string          `mapstructure:"crtPath,omitempty"`
	KeyPath      string          `mapstructure:"keyPath,omitempty"`
	LibPath      string          `mapstructure:"libPath,omitempty"`
	Pass         string          `mapstructure:"pass,omitempty"`
	CrtChainPath string          `mapstructure:"crtChainPath,omitempty"`
	SignData     signer.SignData `mapstructure:"signData"`
}

var (
	config            Config
	signersConfigArr  []signerConfig
	servicesConfigArr []serviceConfig
)

// initConfig reads in config inputFile and ENV variables if set.
func initConfig(cmd *cobra.Command) {
	preParseConfigFlag()

	if configFilePathFlag != "" {
		// Use config inputFile from the flag.
		viper.SetConfigFile(configFilePathFlag)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pdfsigner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pdfsigner")
	}

	viper.AutomaticEnv()      // read in environment variables that match
	viper.SetEnvPrefix("PDF") // set env prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// If a config inputFile is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
		// validate config
		if len(viper.AllSettings()) == 0 {
			log.Fatal(errors.New("config is not properly formatted or empty"))
		}
	}

	if err != nil && configFilePathFlag != "" {
		log.Fatal(err)
	}

	// Unmarshal the full config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Error decoding config:", err)
	}

	// Convert nested config to flat arrays for backward compatibility
	for name, signer := range config.Signers {
		signer.Name = name // Set the name field
		signersConfigArr = append(signersConfigArr, signer)
	}

	for name, service := range config.Services {
		service.Name = name // Set the name field
		servicesConfigArr = append(servicesConfigArr, service)
	}

	licenseStrConfOrFlag = config.LicensePath

	// setup CLI overrides for signers and services of the config if it's multi command
	setupMultiSignersFlags(cmd)
	setupMultiServiceFlags(cmd)
}

// needed to override signers and services config settings.
func preParseConfigFlag() {
	const configFlagName = "--config"

	args := strings.Join(os.Args[1:], " ")
	if strings.Contains(args, configFlagName) {
		fields := strings.Fields(args)
		for i, f := range fields {
			if strings.Contains(f, configFlagName) && len(fields) > i+1 {
				configFilePathFlag = fields[i+1]
			}
		}
	}
}
