package cmd

import (
	"fmt"

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
	// Add RunE to handle the case when only 'serve' is provided with --config
	RunE: func(cmd *cobra.Command, args []string) error {
		// If config file is specified, check if we can determine what to serve
		if len(config.Services) > 0 {
			log.Info("Found signers in config file, running in serve signers mode")
			// Run the serve signers command
			return serveWithMultipleSignersCmd.RunE(cmd, args)
		}

		// If no config is specified, show help for the serve command
		return cmd.Help()
	},
}

// servePEMCmd runs web api with PEM using only flags.
var servePEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Serve using PEM signer",
	RunE: func(cmd *cobra.Command, attr []string) error {
		// loading jobs from the db
		err := signVerifyQueue.LoadFromDB()
		if err != nil {
			return err
		}

		config := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		err = config.SignData.SetPEM(config.Cert, config.Key, config.Chain)
		if err != nil {
			return fmt.Errorf("failed to set PEM certificate data: %w", err)
		}

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(config.SignData)

		return nil
	},
}

// servePKSC11Cmd runs web api with PKSC11 using only flags.
var servePKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Serve using PKSC11 signer",
	RunE: func(cmd *cobra.Command, attr []string) error {
		// loading jobs from the db
		err := signVerifyQueue.LoadFromDB()
		if err != nil {
			return fmt.Errorf("failed to load jobs from the db: %w", err)
		}

		// create signer config
		config := signerConfig{}

		// bind signer flags to config
		bindSignerFlagsToConfig(cmd, &config)

		// set sign data
		err = config.SignData.SetPKSC11(config.Lib, config.Pass, config.Chain)
		if err != nil {
			return fmt.Errorf("failed to set PKSC11 configuration: %w", err)
		}

		// start web api with runners using unnamed signer
		startWebAPIWithRunnersUnnamedSigner(config.SignData)

		return nil
	},
}

// serveWithMultipleSignersCmd runs web api using multiple signers, with NO possibility to override it with flags.
var serveWithMultipleSignersCmd = &cobra.Command{
	Use:   "signers",
	Short: "Serve with multiple signers from the config",
	Long:  `Runs multiple signers. Settings couldn't be overwritten`,
	RunE: func(cmd *cobra.Command, signerNames []string) error {
		// loading jobs from the db
		err := signVerifyQueue.LoadFromDB()
		if err != nil {
			return err
		}

		// Get signers from config if not provided as arguments
		if len(signerNames) < 1 {
			// Use all configured signers
			for name := range config.Signers {
				signerNames = append(signerNames, name)
			}

			// Check again if we have signers
			if len(signerNames) < 1 {
				return fmt.Errorf("no signers found in config and none provided as arguments")
			}
		}

		// setup signers
		for _, sn := range signerNames {
			err = setupSigner(sn)
			if err != nil {
				return fmt.Errorf("failed to setup signer: %w", err)
			}
		}

		// setup verifier
		setupVerifier()

		// start web api with runners
		startWebAPIWithProcessor(signerNames)

		return nil
	},
}

// startWebAPIWithRunnersUnnamedSigner start the web api.
func startWebAPIWithRunnersUnnamedSigner(signData signer.SignData) {
	id := "signer"
	signVerifyQueue.AddSignUnit(id, signData)
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
	// No need to call parseConfigFlag since config is now a global flag
	parseServeFlags(serveWithMultipleSignersCmd)
}
