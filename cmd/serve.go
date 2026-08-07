package cmd

import (
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/signer"
	"github.com/digitorus/pdfsigner/webapi"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run web server to sign and verify files using HTTP protocol",
	Long:  `Web API allows to sign and verify files by communicating with the application using HTTP protocol`,
}

// servePEMCmd runs web api with PEM using only flags.
var servePEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Serve using PEM signer",
	Run: func(cmd *cobra.Command, attr []string) {
		// require license
		err := requireLicense()
		if err != nil {
			log.Fatal(err)
		}

		// loading jobs from the db
		err = signVerifyQueue.LoadFromDB()
		if err != nil {
			log.Fatal(err)
		}

		config := signerConfig{Type: "pem"}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(mustBuildIdentity(config), config.Options)
	},
}

// servePKSC11Cmd runs web api with PKSC11 using only flags.
var servePKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Serve using PKSC11 signer",
	Run: func(cmd *cobra.Command, attr []string) {
		// require license
		err := requireLicense()
		if err != nil {
			log.Fatal(err)
		}

		// loading jobs from the db
		err = signVerifyQueue.LoadFromDB()
		if err != nil {
			log.Fatal(err)
		}

		// create signer config
		config := signerConfig{Type: "pksc11"}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(mustBuildIdentity(config), config.Options)
	},
}

// serveWithMultipleSignersCmd runs web api using multiple signers, with NO possibility to override it with flags.
var serveWithMultipleSignersCmd = &cobra.Command{
	Use:   "signers",
	Short: "Serve with multiple signers from the config",
	Long:  `Runs multiple signers. Settings couldn't be overwritten`,
	Run: func(cmd *cobra.Command, signerNames []string) {
		// require license
		err := requireLicense()
		if err != nil {
			log.Fatal(err)
		}

		// loading jobs from the db
		err = signVerifyQueue.LoadFromDB()
		if err != nil {
			log.Fatal(err)
		}

		// check if the signer names provided
		if len(signerNames) < 1 {
			log.Fatal("signers are not provided")
		}

		// setup signers
		for _, sn := range signerNames {
			setupSigner(sn)
		}

		// setup verifier
		setupVerifier()

		// start web api with runners
		startWebAPIWithProcessor(signerNames)
	},
}

// startWebAPIWithRunnersUnnamedSigner start the web api.
func startWebAPIWithRunnersUnnamedSigner(identity *signer.Identity, opts signer.Options) {
	id := "signer"
	signVerifyQueue.AddSignUnit(id, identity, opts)
	log.Println(signVerifyQueue)
	startWebAPIWithProcessor([]string{id})
}

// startWebAPIWithProcessor.
func startWebAPIWithProcessor(allowedSigners []string) {
	wa := webapi.NewWebAPI(getAddrPort(), signVerifyQueue, allowedSigners, ver, validateSignature)

	// run queue processors
	signVerifyQueue.StartProcessor()

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
	parsePEMCertificateFlags(servePEMCmd)
	parseServeFlags(servePEMCmd)

	// add PKSC11 serve command and parse related flags
	serveCmd.AddCommand(servePKSC11Cmd)
	parseCommonFlags(servePKSC11Cmd)
	parsePKSC11CertificateFlags(servePKSC11Cmd)
	parseServeFlags(servePKSC11Cmd)

	// add serve with multiple signers and parse related flags
	serveCmd.AddCommand(serveWithMultipleSignersCmd)
	parseConfigFlag(serveWithMultipleSignersCmd)
	parseServeFlags(serveWithMultipleSignersCmd)
}
