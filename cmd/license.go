package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/digitorus/pdfsigner/license"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// licenseCmd represents the license command.
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Check and update license",
}

// licenseSetupCmd represents the license setup command.
var licenseSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "license setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		// initialize license
		err := initializeLicense()
		if err != nil {
			return fmt.Errorf("failed to initialize license: %w", err)
		}

		// print license info
		cmd.Print(license.LD.Info())

		return nil
	},
}

// licenseInfoCmd represents the license info command.
var licenseInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "license info",
	RunE: func(cmd *cobra.Command, args []string) error {
		// load license
		err := license.Load()
		if err != nil {
			// try to initialize license with buid-in license
			err := license.Initialize(nil)
			if err != nil {
				return fmt.Errorf("failed to initialize license: %w", err)
			}
		}

		// print license info
		cmd.Print(license.LD.Info())

		return nil
	},
}

func init() {
	RootCmd.AddCommand(licenseCmd)
	licenseCmd.AddCommand(licenseSetupCmd)
	licenseCmd.AddCommand(licenseInfoCmd)

	// Add license path flag
	licenseCmd.PersistentFlags().String("license-path", "", "Path to license file")
	_ = viper.BindPFlag("licensePath", licenseCmd.PersistentFlags().Lookup("license-path"))
}

// initializeLicense loads the license file with provided path from viper config or command line.
func initializeLicense() error {
	// Get license from viper (can be from flag, env var, or config)
	licenseStr := viper.GetString("license")
	licenseFilePath := viper.GetString("licensePath")

	// If neither license string nor license path is provided through viper
	if licenseStr == "" && licenseFilePath == "" {
		// Check if license was provided directly from command line
		if licenseStrConfOrFlag != "" {
			licenseStr = licenseStrConfOrFlag
		} else {
			// As a last resort, get license from stdin
			fmt.Fprint(os.Stdout, "Paste your license here: ")

			var err error
			licenseStr, err = bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read license from stdin: %w", err)
			}
		}
	} else if licenseStr == "" && licenseFilePath != "" {
		// If license path is provided, read from file
		licenseBytes, err := os.ReadFile(licenseFilePath)
		if err != nil {
			return fmt.Errorf("failed to read license file: %w", err)
		}
		licenseStr = string(licenseBytes)
	}

	// Process the license string
	licenseBytes := []byte(strings.Replace(strings.TrimSpace(licenseStr), "\n", "", -1))

	// Initialize license
	err := license.Initialize(licenseBytes)
	if err != nil {
		return fmt.Errorf("failed to initialize license: %w", err)
	}

	return nil
}
