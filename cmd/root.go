package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/digitorus/pdfsigner/version"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile              string
	licenseStrConfOrFlag string
	ver                  version.Version
	licenseValidated     bool // Track if license has been validated
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "pdfsigner",
	Short: "PDFSigner is a multi purpose PDF signer and verifier",
	Long:  `PDFSigner is a multi purpose PDF signer and verifier application it allows to use it as a command line tool and as a watch and sign tool, it also allows to use Web API to sign and verify files and to use multiple services in combination.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip license check for license command and version command
		if cmd.CommandPath() == "pdfsigner license" ||
			cmd.CommandPath() == "pdfsigner license setup" ||
			cmd.CommandPath() == "pdfsigner license info" ||
			cmd.CommandPath() == "pdfsigner version" {
			return nil
		}

		// Check license globally, only once
		if !licenseValidated {
			if err := requireLicense(); err != nil {
				return fmt.Errorf("license error: %w", err)
			}
			log.Debug("License validated successfully")
			licenseValidated = true
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// If config is specified but no subcommand, check if we can determine what to run
		if cfgFile != "" {
			// Config loaded successfully, now check which service to run
			// Check if services are configured in the config file
			if len(config.Services) > 0 {
				log.Debug("Found services in config file, running in services mode")
				// Run the services command
				return multiCmd.RunE(cmd, args)
			}

			// Check if signers are configured without services
			if len(config.Signers) > 0 {
				log.Debug("Found signers in config file, running in serve signers mode")
				// Run the serve signers command
				return serveWithMultipleSignersCmd.RunE(cmd, args)
			}

			return fmt.Errorf("config file provided but no services or signers found")
		}

		// If no config file is specified, show help
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v version.Version) {
	// set version
	ver = v

	// execute root cmd
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// This is called before each command execution
	cobra.OnInitialize(initConfig)

	// Define the config flag as a persistent flag on the root command
	// so it's globally available to all subcommands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pdfsigner)")

	// Add license as a persistent flag as well
	RootCmd.PersistentFlags().StringVar(&licenseStrConfOrFlag, "license", "", "license string")
	_ = viper.BindPFlag("license", RootCmd.PersistentFlags().Lookup("license"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// This is called by cobra.OnInitialize
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Warn("Error finding home directory:", err)
			return
		}

		// Search config in home directory with name ".pdfsigner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pdfsigner")
	}

	viper.SetEnvPrefix("PDF") // set env prefix
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// If a config file was explicitly provided but not found, log an error
			log.Errorf("Error reading config file: %v", err)
		}
		return
	}

	log.Debugf("Using config file: %s", viper.ConfigFileUsed())

	// Load configuration into app-level structures
	if err := loadConfig(); err != nil {
		log.Errorf("Error loading config: %v", err)
		return
	}
}

// loadConfig loads configuration from viper into application-level structures
func loadConfig() error {
	// Load config into global configuration struct
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}
