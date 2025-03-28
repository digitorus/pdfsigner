package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/queues/queue"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	// common flags.
	signerNameFlag           string
	validateSignature        bool
	certificateChainPathFlag string
	inputPathFlag            string
	outputPathFlag           string

	// Signature flags.
	signatureTypeFlag         uint
	docMdpPermsFlag           uint
	signatureInfoNameFlag     string
	signatureInfoLocationFlag string
	signatureInfoReasonFlag   string
	signatureInfoContactFlag  string
	signatureTSAUrlFlag       string
	signatureTSAUsernameFlag  string
	signatureTSAPasswordFlag  string

	// PEM flags.
	certificatePathFlag string
	privateKeyPathFlag  string

	// PKSC11 flags.
	pksc11LibPathFlag string
	pksc11PassFlag    string

	// serve flags.
	serveAddrFlag string
	servePortFlag string
)

// parseCommonFlags binds common flags to variables.
func parseCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().UintVar(&signatureTypeFlag, "type", 1, "Certificate type")
	cmd.PersistentFlags().UintVar(&docMdpPermsFlag, "docmdp", 1, "DocMDP permissions")
	cmd.PersistentFlags().StringVar(&signatureInfoNameFlag, "name", "", "Signature info name")
	cmd.PersistentFlags().StringVar(&signatureInfoLocationFlag, "location", "", "Signature info location")
	cmd.PersistentFlags().StringVar(&signatureInfoReasonFlag, "reason", "", "Signature reason")
	cmd.PersistentFlags().StringVar(&signatureInfoContactFlag, "contact", "", "Signature contact")
	cmd.PersistentFlags().StringVar(&signatureTSAUrlFlag, "tsa-url", "", "TSA url")
	cmd.PersistentFlags().StringVar(&signatureTSAUsernameFlag, "tsa-username", "", "TSA username")
	cmd.PersistentFlags().StringVar(&signatureTSAPasswordFlag, "tsa-password", "", "TSA password")
	cmd.PersistentFlags().StringVar(&certificateChainPathFlag, "chain", "", "Certificate chain path")
	cmd.PersistentFlags().BoolVar(&validateSignature, "validate-signature", true, "Certificate chain path")
}

func parseConfigFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&configFilePathFlag, "config", "", "Path to config file")
	_ = cmd.MarkPersistentFlagRequired("config")
}

// parsePEMCertificateFlags binds PEM specific flags to variables.
func parsePEMCertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&certificatePathFlag, "crt", "", "Path to certificate file")
	cmd.PersistentFlags().StringVar(&privateKeyPathFlag, "key", "", "Path to private key")
}

// parsePKSC11CertificateFlags binds PKSC11 specific flags to variables.
func parsePKSC11CertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&pksc11LibPathFlag, "lib", "", "Path to PKCS11 library")
	cmd.PersistentFlags().StringVar(&pksc11PassFlag, "pass", "", "PKCS11 password")
}

// parseInputPathFlag binds input folder flag to variable.
func parseInputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&inputPathFlag, "in", "", "Input path")
	_ = cmd.MarkPersistentFlagRequired("in")
}

// parseOutputPathFlag binds output folder flag to variable.
func parseOutputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&outputPathFlag, "out", "", "Output path")
	_ = cmd.MarkPersistentFlagRequired("out")
}

// parseSignerName binds signer name flag to variable.
func parseSignerName(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&signerNameFlag, "signer-name", "", "Signer name")
	_ = cmd.MarkPersistentFlagRequired("signer-name")
}

// parseServeFlags binds serve address and port flags to variables.
func parseServeFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&serveAddrFlag, "serve-address", "", "Address to serve Web API")
	_ = cmd.MarkPersistentFlagRequired("serve-address")
	cmd.PersistentFlags().StringVar(&servePortFlag, "serve-port", "", "Port to serve Web API")
	_ = cmd.MarkPersistentFlagRequired("serve-port")
}

func isMultiSignerCmd() bool {
	if len(os.Args) < 3 {
		return false
	}

	cmd := os.Args[1]

	subCmd := os.Args[2]

	return (cmd == "sign" && subCmd == "signer") || (cmd == "serve" && subCmd == "signers") || cmd == "services"
}

// setupMultiSignersFlags setups commands to override signers config settings.
func setupMultiSignersFlags(cmd *cobra.Command) {
	if !isMultiSignerCmd() {
		return
	}

	for i, s := range signersConfigArr {
		// set flagSuffix if multiple signers provided inside config
		flagSuffix := ""
		if len(servicesConfigArr) > 0 {
			flagSuffix = "_" + s.Name
		}

		// set usage suffix
		var usageSuffix string
		if len(servicesConfigArr) > 0 {
			usageSuffix += " " + s.Name
		}

		usageSuffix += " config override flag"

		// create commands with temporary uint variables
		certTypeUint := uint(s.SignData.Signature.CertType)
		docMDPPermUint := uint(s.SignData.Signature.DocMDPPerm)

		cmd.PersistentFlags().UintVar(&certTypeUint, "type"+flagSuffix, certTypeUint, "Certificate type"+usageSuffix)
		cmd.PersistentFlags().UintVar(&docMDPPermUint, "docmdp"+flagSuffix, docMDPPermUint, "DocMDP permissions"+usageSuffix)

		// Store the index for later use in flag handling
		certTypeIndices := make(map[string]int)
		docMDPPermIndices := make(map[string]int)
		certTypeIndices["type"+flagSuffix] = i
		docMDPPermIndices["docmdp"+flagSuffix] = i

		// Add post-processing for these flags
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			for flagName, idx := range certTypeIndices {
				if cmd.PersistentFlags().Changed(flagName) {
					val, _ := cmd.PersistentFlags().GetUint(flagName)
					signersConfigArr[idx].SignData.Signature.CertType = sign.CertType(val)
				}
			}

			for flagName, idx := range docMDPPermIndices {
				if cmd.PersistentFlags().Changed(flagName) {
					val, _ := cmd.PersistentFlags().GetUint(flagName)
					signersConfigArr[idx].SignData.Signature.DocMDPPerm = sign.DocMDPPerm(val)
				}
			}
		}
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.Signature.Info.Name, "name"+flagSuffix, s.SignData.Signature.Info.Name, "Signature info name"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.Signature.Info.Location, "location"+flagSuffix, s.SignData.Signature.Info.Location, "Signature info location"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.Signature.Info.Reason, "reason"+flagSuffix, s.SignData.Signature.Info.Reason, "Signature reason"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.Signature.Info.ContactInfo, "contact"+flagSuffix, s.SignData.Signature.Info.ContactInfo, "Signature contact"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.TSA.URL, "tsa-url"+flagSuffix, s.SignData.TSA.URL, "TSA url"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.TSA.Username, "tsa-username"+flagSuffix, s.SignData.TSA.Username, "TSA username"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].SignData.TSA.Password, "tsa-password"+flagSuffix, s.SignData.TSA.Password, "TSA password"+usageSuffix)
		cmd.PersistentFlags().StringVar(&signersConfigArr[i].CrtChainPath, "chain"+flagSuffix, s.CrtChainPath, "Certificate chain path"+usageSuffix)
	}
}

func isMultiServiceCmd() bool {
	if len(os.Args) < 2 {
		return false
	}

	return os.Args[1] == "services"
}

// setupMultiServiceFlags setups commands to override services config settings.
func setupMultiServiceFlags(cmd *cobra.Command) {
	if !isMultiServiceCmd() {
		return
	}

	for i, s := range servicesConfigArr {
		// set suffix if multiple signers provided inside config
		suffix := ""
		if len(servicesConfigArr) > 0 {
			suffix = "_" + s.Name
		}

		// set usage suffix
		var usageSuffix string
		if len(servicesConfigArr) > 0 {
			usageSuffix += " " + s.Name
		}

		usageSuffix += " config override flag"

		// create commands
		cmd.PersistentFlags().BoolVar(&servicesConfigArr[i].ValidateSignature, "validate-signature"+suffix, true, "Certificate chain path"+usageSuffix)
	}
}

// getAddrPort returns server address and port formatted.
func getAddrPort() string {
	return serveAddrFlag + ":" + servePortFlag
}

// bindSignerFlagsToConfig binds signer specific flags to variables.
// Since viper is not supporting binding flags to an item of the array we use this workaround.
func bindSignerFlagsToConfig(cmd *cobra.Command, c *signerConfig) {
	log.Debug("bindSignerFlagsToConfig")

	// JobSignConfig
	if cmd.PersistentFlags().Changed("docmdp") {
		c.SignData.Signature.DocMDPPerm = sign.DocMDPPerm(docMdpPermsFlag)
	}

	if cmd.PersistentFlags().Changed("type") {
		c.SignData.Signature.CertType = sign.CertType(signatureTypeFlag)
	}

	if cmd.PersistentFlags().Changed("name") {
		c.SignData.Signature.Info.Name = signatureInfoNameFlag
	}

	if cmd.PersistentFlags().Changed("location") {
		c.SignData.Signature.Info.Location = signatureInfoLocationFlag
	}

	if cmd.PersistentFlags().Changed("reason") {
		c.SignData.Signature.Info.Reason = signatureInfoReasonFlag
	}

	if cmd.PersistentFlags().Changed("contact") {
		c.SignData.Signature.Info.ContactInfo = signatureInfoContactFlag
	}

	if cmd.PersistentFlags().Changed("tsa-password") {
		c.SignData.TSA.URL = signatureTSAUrlFlag
	}

	if cmd.PersistentFlags().Changed("tsa-url") {
		c.SignData.TSA.Password = signatureTSAPasswordFlag
	}

	// Certificate chain
	if cmd.PersistentFlags().Changed("chain") {
		c.CrtChainPath = certificateChainPathFlag
	}

	// PEM
	if cmd.PersistentFlags().Changed("crt") {
		c.CrtPath = certificatePathFlag
	}

	if cmd.PersistentFlags().Changed("key") {
		c.KeyPath = privateKeyPathFlag
	}

	// PKSC11
	if cmd.PersistentFlags().Changed("lib") {
		c.LibPath = pksc11LibPathFlag
	}

	if cmd.PersistentFlags().Changed("pass") {
		c.Pass = pksc11PassFlag
	}
}

// getSignerConfigByName returns config of the signer by name.
func getSignerConfigByName(signerName string) signerConfig {
	if signerName == "" {
		log.Fatal("signer name is empty")
	}

	// find signer config
	var s signerConfig
	for _, s = range signersConfigArr {
		if s.Name == signerName {
			return s
		}
	}

	// fail if signer not found
	log.Fatal("signer not found")

	return s
}

// getConfigServiceByName returns service config by name.
func getConfigServiceByName(serviceName string) serviceConfig {
	if serviceName == "" {
		log.Fatal("service name is not provided")
	}

	// find service config
	var s serviceConfig
	for _, s = range servicesConfigArr {
		if s.Name == serviceName {
			return s
		}
	}

	// fail if service not found
	log.Fatal("service not found")

	return s
}

// requireLicense loads license.
func requireLicense() error {
	// load license from the db
	var licenseInitErr error

	licenseLoadErr := license.Load()
	if licenseLoadErr != nil {
		// if the license couldn't be loaded try to initialize it
		licenseInitErr = initializeLicense()
	}

	if licenseInitErr != nil {
		return licenseInitErr
	}

	return nil
}

func getOutputFilePathByInputFilePath(inputFilePath, outputFolderPath string) string {
	_, fileName := filepath.Split(inputFilePath)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return filepath.Join(outputFolderPath, fileNameNoExt+"_signed"+filepath.Ext(fileName))
}
