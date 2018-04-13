package cmd

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/digitorus/pdfsigner/signer"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Since Viper is not supporting access to array, we need to make structures and unmarshal config manually
var signerConfigs []signerConfig

type signerConfig struct {
	Name         string
	Type         string
	CrtPath      string
	KeyPath      string
	LibPath      string
	Pass         string
	CrtChainPath string
	SignData     signer.SignData
}

// used for multiple-services command
var servicesConfig []serviceConfig

type serviceConfig struct {
	Name    string
	Type    string
	Signer  string   // signer name
	Signers []string // signer names to serve with multiple signers
	In      string
	Out     string
	Addr    string
	Port    string
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

	licenseFilePathFlag = viper.GetString("licensePath")
}
