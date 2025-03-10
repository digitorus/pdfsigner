package cmd

import (
	"fmt"
	"os"

	"github.com/digitorus/pdfsigner/files"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/signer"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// watchCmd represents the watch command.
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch folder for new files, sign and put to another folder",
	Long:  `Watch folder for new PDF documents, sign it using PEM or PKSC11 or preconfigured signer`,
}

// watchPEMCmd watches folders and signs files with PEM using flags only.
var watchPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Watch and sign with PEM formatted certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		// create signer config
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		if err := c.SignData.SetPEM(c.Cert, c.Key, c.Chain); err != nil {
			return fmt.Errorf("failed to set PEM certificate data: %w", err)
		}

		// start watch
		if err := startWatch(c.SignData); err != nil {
			return fmt.Errorf("failed to start watch process: %w", err)
		}

		return nil
	},
}

// watchPKSC11Cmd watches folders and signs files with PEM using flags only.
var watchPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Watch and sign with PSKC11",
	RunE: func(cmd *cobra.Command, args []string) error {
		// create signer config
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		err := c.SignData.SetPKSC11(c.Lib, c.Pass, c.Chain)
		if err != nil {
			return fmt.Errorf("failed to set PKSC11 configuration: %w", err)
		}

		// start watch
		if err := startWatch(c.SignData); err != nil {
			return fmt.Errorf("failed to start watch process: %w", err)
		}

		return nil
	},
}

// watchBySignerNameCmd watches folders and signs files using singer from the config with possibility to override it with flags.
var watchBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Watch and sign with preconfigured signer",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get signer config from the config file by name
		signerName := viper.GetString("signerName")
		c, err := getSignerConfigByName(signerName)
		if err != nil {
			return err
		}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		switch c.Type {
		case "pem":
			// set sign data
			if err := c.SignData.SetPEM(c.Cert, c.Key, c.Chain); err != nil {
				return fmt.Errorf("failed to set PEM certificate data: %w", err)
			}
		case "pksc11":
			err := c.SignData.SetPKSC11(c.Lib, c.Pass, c.Chain)
			if err != nil {
				return fmt.Errorf("failed to set PKSC11 configuration: %w", err)
			}
		default:
			return fmt.Errorf("unknown signer type: %s", c.Type)
		}

		// start watch
		if err := startWatch(c.SignData); err != nil {
			return fmt.Errorf("failed to start watch process: %w", err)
		}

		return nil
	},
}

// startWatch starts watcher.
func startWatch(signData signer.SignData) error {
	license.LD.AutoSave()

	// Get input and output paths from viper
	inputPath := viper.GetString("in")
	outputPath := viper.GetString("out")
	validateSig := viper.GetBool("validateSignature")

	// Fallback to flag values if not found in viper
	if inputPath == "" {
		inputPath = inputPathFlag
	}
	if outputPath == "" {
		inputPath = outputPathFlag
	}

	// Check if input and output paths exist
	if _, err := os.Stat(inputPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", inputPath)
		} else {
			return fmt.Errorf("cannot access input directory %s: %w", inputPath, err)
		}
	}

	if _, err := os.Stat(outputPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", outputPath)
		} else {
			return fmt.Errorf("cannot access output directory %s: %w", outputPath, err)
		}
	}

	files.Watch(inputPath, func(filePath string, left int) {
		signedFilePath := getOutputFilePathByInputFilePath(filePath, outputPath)
		if err := signer.SignFile(filePath, signedFilePath, signData, validateSig); err != nil {
			log.Errorln(err)
		}
	})

	return nil
}

func init() {
	RootCmd.AddCommand(watchCmd)
	parseCommonFlags(watchCmd)
	parseInputPathFlag(watchCmd)
	parseOutputPathFlag(watchCmd)
	parsePEMCertificateFlags(watchCmd)
	parsePKSC11CertificateFlags(watchCmd)

	// add PEM sign command and parse related flags
	watchCmd.AddCommand(watchPEMCmd)
	parseCommonFlags(watchPEMCmd)
	parseInputPathFlag(watchPEMCmd)
	parseOutputPathFlag(watchPEMCmd)
	parsePEMCertificateFlags(watchPEMCmd)

	// add PKSC11 sign command and parse related flags
	watchCmd.AddCommand(watchPKSC11Cmd)
	parseCommonFlags(watchPKSC11Cmd)
	parseOutputPathFlag(watchPKSC11Cmd)
	parseInputPathFlag(watchPKSC11Cmd)
	parsePKSC11CertificateFlags(watchPKSC11Cmd)

	// add watch command with signer from config and parse related flags
	watchCmd.AddCommand(watchBySignerNameCmd)
	parseCommonFlags(watchBySignerNameCmd)
	parseInputPathFlag(watchBySignerNameCmd)
	parseOutputPathFlag(watchBySignerNameCmd)
	parsePEMCertificateFlags(watchBySignerNameCmd)
	parsePKSC11CertificateFlags(watchBySignerNameCmd)
}
