package cmd

import (
	"log"

	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/queued_verify"
	"github.com/spf13/cobra"
)

var qSign *queued_sign.QSign
var qVerify *queued_verify.QVerify

func init() {
	qSign = queued_sign.NewQSign()
	qVerify = queued_verify.NewQVerify()
}

var (
	// common flags
	signerNameFlag           string
	certificateChainPathFlag string
	inputPathFlag            string
	outputPathFlag           string

	// Signature flags
	signatureApprovalFlag     bool
	signatureTypeFlag         uint
	signatureInfoNameFlag     string
	signatureInfoLocationFlag string
	signatureInfoReasonFlag   string
	signatureInfoContactFlag  string
	signatureTSAUrlFlag       string
	signatureTSAUsernameFlag  string
	signatureTSAPasswordFlag  string

	// PEM flags
	certificatePathFlag string
	privateKeyPathFlag  string

	// PKSC11 flags
	pksc11LibPathFlag string
	pksc11PassFlag    string

	// serve flags
	serveAddrFlag string
	servePortFlag string
)

func parseCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&signatureApprovalFlag, "approval", false, "Approval")
	cmd.PersistentFlags().UintVar(&signatureTypeFlag, "type", 1, "Certificate type")
	cmd.PersistentFlags().StringVar(&signatureInfoNameFlag, "info-name", "", "Signature info name")
	cmd.PersistentFlags().StringVar(&signatureInfoLocationFlag, "info-location", "", "Signature info location")
	cmd.PersistentFlags().StringVar(&signatureInfoReasonFlag, "info-reason", "", "Signature info reason")
	cmd.PersistentFlags().StringVar(&signatureInfoContactFlag, "info-contact", "", "Signature info contact")
	cmd.PersistentFlags().StringVar(&signatureTSAUrlFlag, "tsa-url", "", "TSA url")
	cmd.PersistentFlags().StringVar(&signatureTSAUsernameFlag, "tsa-username", "", "TSA username")
	cmd.PersistentFlags().StringVar(&signatureTSAPasswordFlag, "tsa-password", "", "TSA password")
	cmd.PersistentFlags().StringVar(&certificateChainPathFlag, "chain", "", "Certificate chain")
}

func parsePEMCertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&certificatePathFlag, "crt", "", "Certificate path")
	cmd.PersistentFlags().StringVar(&privateKeyPathFlag, "key", "", "Private key path")
}

func parsePKSC11CertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&pksc11LibPathFlag, "lib", "", "Path to PKCS11 library")
	cmd.PersistentFlags().StringVar(&pksc11PassFlag, "pass", "", "PKCS11 password")
}

func parseInputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&inputPathFlag, "in", "", "Input path")
	cmd.MarkPersistentFlagRequired("in")
}

func parseOutputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&outputPathFlag, "out", "", "Output path")
	cmd.MarkPersistentFlagRequired("out")
}

func parseSignerName(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&signerNameFlag, "signer-name", "", "Signer name")
	cmd.MarkPersistentFlagRequired("signer-name")
}

func parseServeFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&serveAddrFlag, "serve-address", "", "serve address")
	cmd.MarkPersistentFlagRequired("serve-address")
	cmd.PersistentFlags().StringVar(&servePortFlag, "serve-port", "", "serve port")
	cmd.MarkPersistentFlagRequired("serve-port")
}

func getAddrPort() string {
	return serveAddrFlag + ":" + servePortFlag
}

// Since viper is not supporting binding flags to an item of the array we use this workaround
func bindSignerFlagsToConfig(cmd *cobra.Command, c *signerConfig) {
	// SignData
	if cmd.PersistentFlags().Changed("approval") {
		c.SignData.Signature.Approval = signatureApprovalFlag
	}
	if cmd.PersistentFlags().Changed("type") {
		c.SignData.Signature.CertType = uint32(signatureTypeFlag)
	}
	if cmd.PersistentFlags().Changed("info-name") {
		c.SignData.Signature.Info.Name = signatureInfoNameFlag
	}
	if cmd.PersistentFlags().Changed("info-location") {
		c.SignData.Signature.Info.Location = signatureInfoLocationFlag
	}
	if cmd.PersistentFlags().Changed("info-reason") {
		c.SignData.Signature.Info.Reason = signatureInfoReasonFlag
	}
	if cmd.PersistentFlags().Changed("info-contact") {
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

func getSignerConfigByName(signerName string) signerConfig {
	if signerName == "" {
		log.Fatal("signer name is empty")
	}

	var s signerConfig
	for _, s = range signerConfigs {
		if s.Name == signerName {
			return s
		}
	}
	log.Fatal("signer not found")
	return s
}

func getConfigServiceByName(serviceName string) serviceConfig {
	if serviceName == "" {
		log.Fatal("service name is empty")
	}

	var s serviceConfig
	for _, s = range servicesConfig {
		if s.Name == serviceName {
			return s
		}
	}

	log.Fatal("service not found")
	return s
}

func requireConfig(cmd *cobra.Command) {
	v, err := cmd.Flags().GetString("config")
	if err != nil || v == "" {
		log.Fatal("config is not provided")
	}
}
