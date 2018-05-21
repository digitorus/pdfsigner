package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsigner/signer"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// servicesConfig stores configs for signers for multiple-services command
// Since Viper is not supporting access to array, we need to make structures and unmarshal config manually
var signerConfigs []signerConfig

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
	// SignData contains data needed for signing
	SignData signer.SignData
}

// servicesConfig stores configs for services for multiple-services command
var servicesConfig []serviceConfig

// serviceConfig is a config of the service
type serviceConfig struct {
	// Name is the name of the service
	Name string
	// Type is the type of the service. Watch, Serve.
	Type string

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
	// If a config inputFile is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config inputFile:", viper.ConfigFileUsed())
	}
	if err != nil && RootCmd.Flags().Changed("config") {
		log.Fatal(err)
	}

	// unmarshal signers
	if err := viper.UnmarshalKey("signer", &signerConfigs); err != nil {
		log.Fatal(err)
	}
	// unmarshal services for mixed command
	if err := viper.UnmarshalKey("service", &servicesConfig); err != nil {
		log.Fatal(err)
	}

	// assign licensePath config value to variable
	licenseFilePathFlag = viper.GetString("licensePath")
}
