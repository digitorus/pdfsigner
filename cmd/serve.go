package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/signer"
	"bitbucket.org/digitorus/pdfsigner/webapi"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Sign command",
	Long:  `Long multiline description here`,
}

var servePEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Sign PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, attr []string) {
		config := signerConfig{}
		bindSignerFlagsToConfig(cmd, &config)
		config.SignData.SetPEM(config.CrtPath, config.KeyPath, config.CrtChainPath)

		serveWithUnnamedSigner(config.SignData)
	},
}

var servePKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Signs PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, attr []string) {
		config := signerConfig{}
		bindSignerFlagsToConfig(cmd, &config)
		config.SignData.SetPKSC11(config.LibPath, config.Pass, config.CrtChainPath)

		serveWithUnnamedSigner(config.SignData)
	},
}

var serveWithSingleSignerCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with serve from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, attr []string) {
		// get config signer by name
		config := getSignerConfigByName(signerNameFlag)
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		switch config.Type {
		case "pem":
			config.SignData.SetPEM(config.CrtPath, config.KeyPath, config.CrtChainPath)
		case "pksc11":
			config.SignData.SetPKSC11(config.LibPath, config.Pass, config.CrtChainPath)
		}

		qSign.AddSigner(signerNameFlag, config.SignData, 10)

		serve()
	},
}

var serveWithMultipleSignersCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with serve from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, signerNames []string) {
		for _, sn := range signerNames {
			setupSigner(sn)
		}
		serve()
	},
}

func serveWithUnnamedSigner(signData signer.SignData) {
	// TODO: hash something from signer to identify completed jobs inside db
	id := "someid"
	qSign.AddSigner(id, signData, 10)
	serve()
}

func serve() {
	wa := webapi.NewWebAPI(getAddrPort(), qSign, []string{})
	wa.Serve()
}

func init() {
	RootCmd.AddCommand(serveCmd)

	//PEM serve command
	serveCmd.AddCommand(servePEMCmd)
	parseCommonFlags(servePEMCmd)
	parseInputPathFlag(servePEMCmd)
	parseOutputPathFlag(servePEMCmd)
	parsePEMCertificateFlags(servePEMCmd)
	parseServeFlags(servePEMCmd)

	//PKSC11 serve command
	serveCmd.AddCommand(servePKSC11Cmd)
	parseCommonFlags(servePKSC11Cmd)
	parseInputPathFlag(servePKSC11Cmd)
	parseOutputPathFlag(servePKSC11Cmd)
	parsePKSC11CertificateFlags(servePKSC11Cmd)
	parseServeFlags(servePKSC11Cmd)

	// serve with serve from config inputFile
	serveCmd.AddCommand(serveWithSingleSignerCmd)
	parseSignerName(serveWithSingleSignerCmd)
	parseInputPathFlag(serveWithSingleSignerCmd)
	parseOutputPathFlag(serveWithSingleSignerCmd)
	parsePEMCertificateFlags(serveWithSingleSignerCmd)
	parsePKSC11CertificateFlags(serveWithSingleSignerCmd)
	parseServeFlags(serveWithSingleSignerCmd)

	// serve with serve from config inputFile
	serveCmd.AddCommand(serveWithMultipleSignersCmd)
	parseServeFlags(serveWithMultipleSignersCmd)

}
