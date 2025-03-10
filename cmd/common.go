package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/queues/queue"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// signVerifyQueue stores queue for signs.
var signVerifyQueue *queue.Queue

func init() {
	// initialize queues
	signVerifyQueue = queue.NewQueue()
}

func setupVerifier() {
	signVerifyQueue.AddVerifyUnit()
}

var (
	// Common flag variables - now only used for command-line arguments that don't map directly to config
	validateSignature bool
	inputPathFlag     string
	outputPathFlag    string
)

// parseCommonFlags binds common flags to variables.
func parseCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().UintP("type", "t", 1, "Certificate type")
	_ = viper.BindPFlag("type", cmd.PersistentFlags().Lookup("type"))

	cmd.PersistentFlags().UintP("docmdp", "d", 1, "DocMDP permissions")
	_ = viper.BindPFlag("docmdp", cmd.PersistentFlags().Lookup("docmdp"))

	cmd.PersistentFlags().StringP("signer", "s", "", "Name of signer configuration")
	_ = viper.BindPFlag("signer", cmd.PersistentFlags().Lookup("signer"))

	cmd.PersistentFlags().StringP("name", "n", "", "Signature info name")
	_ = viper.BindPFlag("name", cmd.PersistentFlags().Lookup("name"))

	cmd.PersistentFlags().StringP("location", "l", "", "Signature info location")
	_ = viper.BindPFlag("location", cmd.PersistentFlags().Lookup("location"))

	cmd.PersistentFlags().StringP("reason", "r", "", "Signature reason")
	_ = viper.BindPFlag("reason", cmd.PersistentFlags().Lookup("reason"))

	cmd.PersistentFlags().StringP("contact", "c", "", "Signature contact")
	_ = viper.BindPFlag("contact", cmd.PersistentFlags().Lookup("contact"))

	cmd.PersistentFlags().String("tsa-url", "", "TSA url")
	_ = viper.BindPFlag("tsa.url", cmd.PersistentFlags().Lookup("tsa-url"))

	cmd.PersistentFlags().String("tsa-username", "", "TSA username")
	_ = viper.BindPFlag("tsa.username", cmd.PersistentFlags().Lookup("tsa-username"))

	cmd.PersistentFlags().String("tsa-password", "", "TSA password")
	_ = viper.BindPFlag("tsa.password", cmd.PersistentFlags().Lookup("tsa-password"))

	cmd.PersistentFlags().BoolVar(&validateSignature, "validate-signature", true, "Validate signature")
	_ = viper.BindPFlag("validateSignature", cmd.PersistentFlags().Lookup("validate-signature"))
}

// parsePEMCertificateFlags binds PEM specific flags to variables.
func parsePEMCertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("cert", "", "Path to certificate file")
	_ = viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert"))

	cmd.PersistentFlags().String("key", "", "Path to private key")
	_ = viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key"))

	cmd.PersistentFlags().String("chain", "", "Certificate chain path")
	_ = viper.BindPFlag("chain", cmd.PersistentFlags().Lookup("chain"))
}

// parsePKSC11CertificateFlags binds PKSC11 specific flags to variables.
func parsePKSC11CertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("lib", "", "Path to PKCS11 library")
	_ = viper.BindPFlag("lib", cmd.PersistentFlags().Lookup("lib"))

	cmd.PersistentFlags().String("pass", "", "PKCS11 password")
	_ = viper.BindPFlag("pass", cmd.PersistentFlags().Lookup("pass"))
}

// parseInputPathFlag binds input folder flag to variable.
func parseInputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&inputPathFlag, "in", "", "Input path")
	_ = cmd.MarkPersistentFlagRequired("in")
	_ = viper.BindPFlag("in", cmd.PersistentFlags().Lookup("in"))
}

// parseOutputPathFlag binds output folder flag to variable.
func parseOutputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&outputPathFlag, "out", "", "Output path")
	_ = cmd.MarkPersistentFlagRequired("out")
	_ = viper.BindPFlag("out", cmd.PersistentFlags().Lookup("out"))
}

// parseServeFlags binds serve address and port flags to variables.
func parseServeFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("serve-address", "", "Address to serve Web API")
	_ = cmd.MarkPersistentFlagRequired("serve-address")
	_ = viper.BindPFlag("serve.address", cmd.PersistentFlags().Lookup("serve-address"))

	cmd.PersistentFlags().String("serve-port", "", "Port to serve Web API")
	_ = cmd.MarkPersistentFlagRequired("serve-port")
	_ = viper.BindPFlag("serve.port", cmd.PersistentFlags().Lookup("serve-port"))
}

// getAddrPort returns server address and port formatted.
func getAddrPort() string {
	addr := viper.GetString("serve.address")
	port := viper.GetString("serve.port")
	return addr + ":" + port
}

// bindSignerFlagsToConfig populates signer config with values from config file,
// then overrides with any explicitly provided command line flags.
func bindSignerFlagsToConfig(cmd *cobra.Command, c *signerConfig) {
	// First set values from the configuration file if available
	if c.SignData.Signature.DocMDPPerm == 0 {
		c.SignData.Signature.DocMDPPerm = sign.DocMDPPerm(viper.GetUint("docmdp"))
	}

	if c.SignData.Signature.CertType == 0 {
		c.SignData.Signature.CertType = sign.CertType(viper.GetUint("type"))
	}

	if c.SignData.Signature.Info.Name == "" {
		c.SignData.Signature.Info.Name = viper.GetString("name")
	}

	if c.SignData.Signature.Info.Location == "" {
		c.SignData.Signature.Info.Location = viper.GetString("location")
	}

	if c.SignData.Signature.Info.Reason == "" {
		c.SignData.Signature.Info.Reason = viper.GetString("reason")
	}

	if c.SignData.Signature.Info.ContactInfo == "" {
		c.SignData.Signature.Info.ContactInfo = viper.GetString("contact")
	}

	if c.SignData.TSA.URL == "" {
		c.SignData.TSA.URL = viper.GetString("tsa.url")
	}

	if c.SignData.TSA.Password == "" {
		c.SignData.TSA.Password = viper.GetString("tsa.password")
	}

	if c.SignData.TSA.Username == "" {
		c.SignData.TSA.Username = viper.GetString("tsa.username")
	}

	if c.Chain == "" {
		c.Chain = viper.GetString("chain")
	}

	if c.Cert == "" {
		c.Cert = viper.GetString("cert")
	}

	if c.Key == "" {
		c.Key = viper.GetString("key")
	}

	if c.Lib == "" {
		c.Lib = viper.GetString("lib")
	}

	if c.Pass == "" {
		c.Pass = viper.GetString("pass")
	}

	// Override with command line flags only if explicitly provided
	// Define flag mappings to their corresponding setters
	stringFlags := map[string]func(string){
		"name":         func(val string) { c.SignData.Signature.Info.Name = val },
		"location":     func(val string) { c.SignData.Signature.Info.Location = val },
		"reason":       func(val string) { c.SignData.Signature.Info.Reason = val },
		"contact":      func(val string) { c.SignData.Signature.Info.ContactInfo = val },
		"tsa-url":      func(val string) { c.SignData.TSA.URL = val },
		"tsa-password": func(val string) { c.SignData.TSA.Password = val },
		"tsa-username": func(val string) { c.SignData.TSA.Username = val },
		"chain":        func(val string) { c.Chain = val },
		"cert":         func(val string) { c.Cert = val },
		"key":          func(val string) { c.Key = val },
		"lib":          func(val string) { c.Lib = val },
		"pass":         func(val string) { c.Pass = val },
	}

	uintFlags := map[string]func(uint){
		"docmdp": func(val uint) { c.SignData.Signature.DocMDPPerm = sign.DocMDPPerm(val) },
		"type":   func(val uint) { c.SignData.Signature.CertType = sign.CertType(val) },
	}

	// Process string flags
	for flagName, setter := range stringFlags {
		if cmd.Flags().Changed(flagName) {
			val, _ := cmd.Flags().GetString(flagName)
			setter(val)
		}
	}

	// Process uint flags
	for flagName, setter := range uintFlags {
		if cmd.Flags().Changed(flagName) {
			val, _ := cmd.Flags().GetUint(flagName)
			setter(val)
		}
	}
}

// requireLicense loads license.
func requireLicense() error {
	// load license from the db
	var licenseInitErr error

	licenseLoadErr := license.Load()
	if licenseLoadErr != nil {
		// try to initialize license with buid-in license
		err := license.Initialize(nil)
		if err != nil {
			// if the license couldn't be loaded try to initialize it
			licenseInitErr = initializeLicense()
		}
	}

	if licenseInitErr != nil {
		return fmt.Errorf("license error: %w", licenseInitErr)
	}

	return nil
}

func getOutputFilePathByInputFilePath(inputFilePath, outputFolderPath string) string {
	_, fileName := filepath.Split(inputFilePath)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return filepath.Join(outputFolderPath, fileNameNoExt+"_signed"+filepath.Ext(fileName))
}
