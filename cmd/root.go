package cmd

import (
	"fmt"
	"os"

	"github.com/digitorus/pdfsigner/version"
	"github.com/spf13/cobra"
)

// configFilePathFlag contains path to config file.
var configFilePathFlag string

// licenseStrConfOrFlag contains path to license file.
var licenseStrConfOrFlag string

var ver version.Version

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "pdfsigner",
	Short: "PDFSigner is a multi purpose PDF signer and verifier",
	Long:  `PDFSigner is a multi purpose PDF signer and verifier application it allows to use it as a command line tool and as a watch and sign tool, it also allows to use Web API to sign and verify files and to use multiple services in combination.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v version.Version) {
	// parse config flag to make it available before cobra
	initConfig(RootCmd)

	// set version
	ver = v

	// execute root cmd
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// set the log flags
	cobra.OnInitialize()

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// RootCmd.PersistentFlags().StringVar(&configFilePathFlag, "config", "", "")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	RootCmd.PersistentFlags().StringVar(&licenseStrConfOrFlag, "license", "", "license string")
}
