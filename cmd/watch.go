package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/files"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch folder for new files, sign and put to another folder",
	Long:  `Watch folder for new PDF documents, sign it using PEM or PKSC11 or preconfigured signer`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return requireLicense()
	},
}

// watchPEMCmd watches folders and signs files with PEM using flags only
var watchPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Watch and sign with PEM formatted certificate",
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)

		// start watch
		startWatch(c.SignData)
	},
}

// watchPKSC11Cmd watches folders and signs files with PEM using flags only
var watchPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Watch and sign with PSKC11",
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &c)

		// set sign data
		c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)

		// start watch
		startWatch(c.SignData)
	},
}

// watchBySignerNameCmd wathces folders and signs files using singer from the config with possibility to override it with flags
var watchBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Watch and sign with preconfigured signer",
	Run: func(cmd *cobra.Command, args []string) {
		// get signer config from the config file by name
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

		// start watch
		startWatch(c.SignData)
	},
}

// startWatch starts watcher
func startWatch(signData signer.SignData) {
	license.LD.AutoSave()

	files.Watch(inputPathFlag, func(filePath string, left int) {
		signer.SignFile(filePath, outputPathFlag, signData, validateSignature)
	})
}

func init() {
	RootCmd.AddCommand(watchCmd)

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

	// add watch command with signer signer from config and parse related flags
	watchCmd.AddCommand(watchBySignerNameCmd)
	parseConfigFlag(watchBySignerNameCmd)
	parseSignerName(watchBySignerNameCmd)
	parseCommonFlags(watchBySignerNameCmd)
	parseInputPathFlag(watchBySignerNameCmd)
	parseOutputPathFlag(watchBySignerNameCmd)
	parsePEMCertificateFlags(watchBySignerNameCmd)
	parsePKSC11CertificateFlags(watchBySignerNameCmd)
}
