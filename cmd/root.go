package cmd

import (
	"fmt"
	"os"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/spf13/cobra"
)

// configFilePathFlag contains path to config file
var configFilePathFlag string

// licenseFilePathFlag contains path to license file
var licenseFilePathFlag string

var ver version.Version

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "pdfsigner",
	Short: "A brief description of your application",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// load license from the db
		err := license.Load()
		if err != nil {
			// if the license couldn't be loaded try to initialize it
			return initializeLicense()
		}

		return nil
	},
	Long: `Long multiline description`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v version.Version) {
	ver = v
	//RootCmd.SetArgs(os.Args[1:2])
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// set the log flags
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&configFilePathFlag, "config", "", "")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	RootCmd.PersistentFlags().StringVar(&licenseFilePathFlag, "license", "", "license file path")
}
