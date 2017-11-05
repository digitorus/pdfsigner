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
var signersConfig []signerConfig

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

// used for mixed command
var servicesConfig []serviceConfig

type serviceConfig struct {
	Name   string
	Type   string
	Signer string
	In     string
	Out    string
	Port   int
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFilePathFlag != "" {
		// Use config file from the flag.
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

	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// unmarshal signers
	if err := viper.UnmarshalKey("signer", &signersConfig); err != nil {
		log.Fatal(err)
	}
	// unmarshal services for mixed command
	if err := viper.UnmarshalKey("service", &servicesConfig); err != nil {
		log.Fatal(err)
	}
}