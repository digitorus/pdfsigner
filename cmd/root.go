package cmd

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/spf13/cobra"
)

var configFilePathFlag string
var ver version.Version

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "pdfsigner",
	Short: "A brief description of your application",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := license.Load()
		if err != nil {
			return initializeLicense()
		}
		return nil
	},
	Long: `Long multiline description`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//      Run: func(cmd *cobra.Command, args []string) { },
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&configFilePathFlag, "config", "", "")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
