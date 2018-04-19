package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/files"
	"github.com/spf13/cobra"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign command",
	Long:  `Long multiline description here`,
}

// signPEMCmd signs files with PEM using flags only
var signPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Sign PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)

		// sign files
		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

// signPKSC11Cmd signs files with PKSC11 using flags only
var signPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Signs PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)

		// sign files
		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

// signBySignerNameCmd signs files using singer from the config with possibility to override it with flags
var signBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with signer from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		requireConfig(cmd)

		// find signer config from config file by name
		c := getSignerConfigByName(signerNameFlag)

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		switch c.Type {
		case "pem":
			c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		case "pksc11":
			c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		}

		// sign files
		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

func init() {
	RootCmd.AddCommand(signCmd)

	// add PEM sign command and parse related flags
	signCmd.AddCommand(signPEMCmd)
	parseCommonFlags(signPEMCmd)
	parseOutputPathFlag(signPEMCmd)
	parsePEMCertificateFlags(signPEMCmd)

	// add PKSC11 sign command and parse related flags
	signCmd.AddCommand(signPKSC11Cmd)
	parseCommonFlags(signPKSC11Cmd)
	parseOutputPathFlag(signPKSC11Cmd)
	parsePKSC11CertificateFlags(signPKSC11Cmd)

	// add sign with signer from config command and parse related flags
	signCmd.AddCommand(signBySignerNameCmd)
	parseSignerName(signBySignerNameCmd)
	parseOutputPathFlag(signBySignerNameCmd)
	parsePEMCertificateFlags(signBySignerNameCmd)
	parsePKSC11CertificateFlags(signBySignerNameCmd)
}
