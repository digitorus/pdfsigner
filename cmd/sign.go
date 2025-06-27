package cmd

import (
	"fmt"

	"github.com/digitorus/pdfsigner/files"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// signCmd represents the sign command.
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign files using PEM or PKSC11",
	Long:  `Command line signer allows to sign document using PEM or PKSC11 provided directly as well as using preconfigured signer from the config file.`,
	// Add RunE to handle the case when only 'sign' is provided with --config
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(config.Signers) > 0 {
			return signBySignerNameCmd.RunE(cmd, args)
		}

		return fmt.Errorf("no signers configured")
	},
}

// signPEMCmd signs files with PEM using flags only.
var signPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Sign PDF with PEM formatted certificate",
	RunE: func(cmd *cobra.Command, filePatterns []string) error {
		// require file patterns
		if err := requireFilePatterns(filePatterns); err != nil {
			return err
		}

		// initialize config
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		err := c.SignData.SetPEM(c.Cert, c.Key, c.Chain)
		if err != nil {
			return err
		}

		// optional output directory
		out, _ := cmd.Flags().GetString("out")

		// optional validation of the signature
		validateSignature, _ := cmd.Flags().GetBool("validateSignature")

		// sign files
		return files.SignFilesByPatterns(filePatterns, c.SignData, validateSignature, out)
	},
}

// signPKSC11Cmd signs files with PKSC11 using flags only.
var signPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Signs PDF with PSKC11",
	RunE: func(cmd *cobra.Command, filePatterns []string) error {
		// require file patterns
		if err := requireFilePatterns(filePatterns); err != nil {
			return err
		}

		// initialize config
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		if err := c.SignData.SetPKSC11(c.Lib, c.Pass, c.Chain); err != nil {
			return err
		}

		// optional output directory
		out, _ := cmd.Flags().GetString("out")

		// optional validation of the signature
		validateSignature, _ := cmd.Flags().GetBool("validateSignature")

		// sign files
		return files.SignFilesByPatterns(filePatterns, c.SignData, validateSignature, out)
	},
}

// signBySignerNameCmd signs files using singer from the config with possibility to override it with flags.
var signBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Sign PDF with preconfigured signer",
	RunE: func(cmd *cobra.Command, filePatterns []string) error {
		// require file patterns
		if err := requireFilePatterns(filePatterns); err != nil {
			return err
		}

		// find signer config from config file by name
		signer, _ := cmd.Flags().GetString("signer")
		c, err := getSignerConfigByName(signer)
		if err != nil {
			return err
		}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		switch c.Type {
		case "pksc11":
			err = c.SignData.SetPKSC11(c.Lib, c.Pass, c.Chain)
		default:
			err = c.SignData.SetPEM(c.Cert, c.Key, c.Chain)
		}
		if err != nil {
			return err
		}

		// optional output directory
		out, _ := cmd.Flags().GetString("out")

		// optional validation of the signature
		validateSignature, _ := cmd.Flags().GetBool("validateSignature")

		// sign files
		return files.SignFilesByPatterns(filePatterns, c.SignData, validateSignature, out)
	},
}

func init() {
	RootCmd.AddCommand(signCmd)
	parseCommonFlags(signCmd)
	parsePEMCertificateFlags(signCmd)
	parsePKSC11CertificateFlags(signCmd)

	// add PEM sign command and parse related flags
	signCmd.AddCommand(signPEMCmd)
	parseCommonFlags(signPEMCmd)
	parsePEMCertificateFlags(signPEMCmd)

	// add PKSC11 sign command and parse related flags
	signCmd.AddCommand(signPKSC11Cmd)
	parseCommonFlags(signPKSC11Cmd)
	parsePKSC11CertificateFlags(signPKSC11Cmd)

	// add sign with signer from config command and parse related flags
	signCmd.AddCommand(signBySignerNameCmd)
	parseCommonFlags(signBySignerNameCmd)
	parsePEMCertificateFlags(signBySignerNameCmd)
	parsePKSC11CertificateFlags(signBySignerNameCmd)

	// Add parseOutputDirectoryFlag calls to all commands
	parseOutputDirectoryFlag(signCmd)
	parseOutputDirectoryFlag(signPEMCmd)
	parseOutputDirectoryFlag(signPKSC11Cmd)
	parseOutputDirectoryFlag(signBySignerNameCmd)
}

// requireFilePatterns checks if the filePatterns were provided.
func requireFilePatterns(filePatterns []string) error {
	if len(filePatterns) < 1 {
		return fmt.Errorf("no file patterns provided")
	}
	return nil
}

// parseOutputDirectoryFlag adds the output directory flag to a command
func parseOutputDirectoryFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("out", "o", "", "output directory for signed files (default: same directory as input)")
	_ = viper.BindPFlag("out", cmd.Flags().Lookup("out"))
}
