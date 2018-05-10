package cmd

import (
	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"bitbucket.org/digitorus/pdfsigner/webapi"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve Web API",
	Long:  `Long multiline description here`,
}

// servePEMCmd runs web api with PEM using only flags
var servePEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Serve using PEM signer",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, attr []string) {
		config := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		config.SignData.SetPEM(config.CrtPath, config.KeyPath, config.CrtChainPath)

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(config.SignData)
	},
}

// servePKSC11Cmd runs web api with PKSC11 using only flags
var servePKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Serve using PKSC11 signer",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, attr []string) {
		config := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		config.SignData.SetPKSC11(config.LibPath, config.Pass, config.CrtChainPath)

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(config.SignData)
	},
}

// serveWithSingleSignerCmd runs web api using single signer from the config with possibility to override it with flags
var serveWithSingleSignerCmd = &cobra.Command{
	Use:   "single-signer",
	Short: "Serve with single signer from the config. Overrides settings with CLI",
	Long:  `It allows to run signer from the config and override it's settings'`,
	Run: func(cmd *cobra.Command, attr []string) {
		// check if the config flag is provided
		requireConfig(cmd)

		// get config signer by name
		config := getSignerConfigByName(signerNameFlag)

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		switch config.Type {
		case "pem":
			config.SignData.SetPEM(config.CrtPath, config.KeyPath, config.CrtChainPath)
		case "pksc11":
			config.SignData.SetPKSC11(config.LibPath, config.Pass, config.CrtChainPath)
		}

		// add signer to the queue signers pool
		signVerifyQueue.AddSignUnit(signerNameFlag, config.SignData)

		// start web api with runners
		startWebAPIWithRunners()
	},
}

// serveWithMultipleSignersCmd runs web api using multiple signers, with NO possibility to override it with flags
var serveWithMultipleSignersCmd = &cobra.Command{
	Use:   "multiple-signers",
	Short: "Serve with multiple signers from the config",
	Long:  `It runs multiple signers. Settings couldn't be overwritten'`,
	Run: func(cmd *cobra.Command, signerNames []string) {
		// check if the config flag is provided
		requireConfig(cmd)

		// check if the signer names provided
		if len(signerNames) < 1 {
			log.Fatal("signers are not provided")
		}

		// setup signers
		for _, sn := range signerNames {
			setupSigner(sn)
		}

		// start web api with runners
		startWebAPIWithRunners()
	},
}

// startWebAPIWithRunnersUnnamedSigner start the web api
func startWebAPIWithRunnersUnnamedSigner(signData signer.SignData) {
	id := "unnamed-signer"
	signVerifyQueue.AddSignUnit(id, signData)
	startWebAPIWithRunners()
}

// startWebAPIWithRunners
func startWebAPIWithRunners() {
	wa := webapi.NewWebAPI(getAddrPort(), signVerifyQueue, []string{}, ver)

	// run queue runners
	signVerifyQueue.Runner()

	// run license auto save
	license.LD.AutoSave()

	// run serve
	wa.Serve()
}

func init() {
	RootCmd.AddCommand(serveCmd)

	// add PEM serve command and parse related flags
	serveCmd.AddCommand(servePEMCmd)
	parseCommonFlags(servePEMCmd)
	parseInputPathFlag(servePEMCmd)
	parseOutputPathFlag(servePEMCmd)
	parsePEMCertificateFlags(servePEMCmd)
	parseServeFlags(servePEMCmd)

	// add PKSC11 serve command and parse related flags
	serveCmd.AddCommand(servePKSC11Cmd)
	parseCommonFlags(servePKSC11Cmd)
	parseInputPathFlag(servePKSC11Cmd)
	parseOutputPathFlag(servePKSC11Cmd)
	parsePKSC11CertificateFlags(servePKSC11Cmd)
	parseServeFlags(servePKSC11Cmd)

	// add serve with config single signer and parse related flags
	serveCmd.AddCommand(serveWithSingleSignerCmd)
	parseSignerName(serveWithSingleSignerCmd)
	parseInputPathFlag(serveWithSingleSignerCmd)
	parseOutputPathFlag(serveWithSingleSignerCmd)
	parsePEMCertificateFlags(serveWithSingleSignerCmd)
	parsePKSC11CertificateFlags(serveWithSingleSignerCmd)
	parseServeFlags(serveWithSingleSignerCmd)

	// add serve with multiple signers and parse related flags
	serveCmd.AddCommand(serveWithMultipleSignersCmd)
	parseServeFlags(serveWithMultipleSignersCmd)
}
