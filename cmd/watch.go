package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch command",
	Long:  `Long multiline description here`,
}

var watchPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Watch PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}
		x(cmd, &c)
		c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		runSingleCmdWatch(c.SignData)
	},
}

var watchPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Watch PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}
		bindSignerFlagsToConfig(cmd, &c)
		c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		runSingleCmdWatch(c.SignData)
	},
}

var watchBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with signer from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := getConfigSignerByName(signerNameFlag)
		bindSignerFlagsToConfig(cmd, &c)

		switch c.Type {
		case "pem":
			c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		case "pksc11":
			c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		}

		runSingleCmdWatch(c.SignData)

	},
}

func runSingleCmdWatch(signData signer.SignData) {
	watchFolder := viper.GetString("in")
	outputFolder := viper.GetString("out")
	watch(signData, watchFolder, outputFolder)
}

func init() {
	RootCmd.AddCommand(watchCmd)

	//PEM watch command
	watchCmd.AddCommand(watchPEMCmd)
	parseCommonFlags(watchPEMCmd)
	parseInputPathFlag(watchPEMCmd)
	parseOutputPathFlag(watchPEMCmd)
	parsePEMCertificateFlags(watchPEMCmd)

	//PKSC11 watch command
	watchCmd.AddCommand(watchPKSC11Cmd)
	parseCommonFlags(watchPKSC11Cmd)
	parseOutputPathFlag(watchPKSC11Cmd)
	parseInputPathFlag(watchPKSC11Cmd)
	parsePKSC11CertificateFlags(watchPKSC11Cmd)

	// watch with watcher from config file
	watchCmd.AddCommand(watchBySignerNameCmd)
	parseSignerName(watchBySignerNameCmd)
	parseCommonFlags(watchBySignerNameCmd)
	parseInputPathFlag(watchBySignerNameCmd)
	parseOutputPathFlag(watchBySignerNameCmd)
	parsePEMCertificateFlags(watchBySignerNameCmd)
	parsePKSC11CertificateFlags(watchBySignerNameCmd)
}
