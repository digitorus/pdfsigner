package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// servicesConfigArr stores configs for signers for multiple-services command
// Since Viper is not supporting access to array, we need to make structures and unmarshal config manually
var signersConfigArr []signerConfig

// serviceConfig is a config of the service
type signerConfig struct {
	// Name is the name of the signer used to identify signer inside the config
	Name string
	// Type is the type of the signer. PEM, PKSC11.
	Type string
	// CrtPath is the path to the certificate file.
	CrtPath string
	// KeyPath is the path to the private key file
	KeyPath string
	// LibPath is the path to the PKSC11 library.
	LibPath string
	// Pass is the password for PSKC11.
	Pass string
	// CrtChainPath is the path to chain of the certificates
	CrtChainPath string
	// SignConfig contains data needed for signing
	SignData signer.SignData
}

// servicesConfigArr stores configs for services for multiple-services command
var servicesConfigArr []serviceConfig

// serviceConfig is a config of the service
type serviceConfig struct {
	// Name is the name of the service
	Name string
	// Type is the type of the service. Watch, Serve.
	Type string
	// ValidateSignature allows to verify signature after sign used as default for serve services
	ValidateSignature bool

	// Watch config.
	// Signer is the signer name to be used by watch.
	Signer string
	// In is the input folder used by watch.
	In string
	// Out is the output folder used by watch.
	Out string

	// Serve config.
	// Signers is the signers names to be used by serve.
	Signers []string
	// Addr is the address on which to serve
	Addr string
	// Port is the port on which to serve
	Port string
}

// initConfig reads in config inputFile and ENV variables if set.
func initConfig() {
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
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal signers
	if err := unmarshalSigners(); err != nil {
		log.Fatal(err)
	}

	// unmarshal services for mixed command
	if err := unmarshalServices(); err != nil {
		log.Fatal(err)
	}

	// assign licensePath config value to variable
	licenseStrConfOrFlag = viper.GetString("license")
}

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

func unmarshalSigners() error {
	for key, _ := range viper.AllSettings() {
		if strings.HasPrefix(key, "signer") {
			var sc signerConfig
			if err := viper.UnmarshalKey(key, &sc); err != nil {
				return err
			}
			//sc.Name = strings.Replace(key, "signer", "", 1)
			sc.Name = key
			signersConfigArr = append(signersConfigArr, sc)
		}
	}

	return nil
}

func unmarshalServices() error {
	for key, _ := range viper.AllSettings() {
		if strings.HasPrefix(key, "service") {
			var sc serviceConfig
			if err := viper.UnmarshalKey(key, &sc); err != nil {
				return err
			}
			//sc.Name = strings.Replace(key, "service", "", 1)
			sc.Name = key
			servicesConfigArr = append(servicesConfigArr, sc)
		}
	}
	return nil
}
