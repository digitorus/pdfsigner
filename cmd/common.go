package cmd

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func storeTempFile(file io.Reader) (string, error) {
	// TODO: Should we encrypt temporary files?
	tmpFile, err := ioutil.TempFile("", "pdf")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func findFilesByPatterns(patterns []string) (matchedFiles []string, err error) {
	for _, f := range patterns {
		m, err := filepath.Glob(f)
		if err != nil {
			return matchedFiles, err
		}
		matchedFiles = append(matchedFiles, m...)
	}
	return matchedFiles, err
}

var (
	signatureApproval     bool
	signatureType         uint
	signatureInfoName     string
	signatureInfoLocation string
	signatureInfoReason   string
	signatureInfoContact  string
	signatureTSAUrl       string
	signatureTSAUsername  string
	signatureTSAPassword  string
	signerNameFlag        string
	certificateChainPath  string
	certificatePath       string
	privateKeyPath        string
	pksc11LibPath         string
	pksc11Pass            string
)

func parseCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&signatureApproval, "approval", false, "Approval")
	cmd.PersistentFlags().UintVar(&signatureType, "type", 1, "Certificate type")
	cmd.PersistentFlags().StringVar(&signatureInfoName, "info-name", "", "Signature info name")
	cmd.PersistentFlags().StringVar(&signatureInfoLocation, "info-location", "", "Signature info location")
	cmd.PersistentFlags().StringVar(&signatureInfoReason, "info-reason", "", "Signature info reason")
	cmd.PersistentFlags().StringVar(&signatureInfoContact, "info-contact", "", "Signature info contact")
	cmd.PersistentFlags().StringVar(&signatureTSAUrl, "tsa-url", "", "TSA url")
	cmd.PersistentFlags().StringVar(&signatureTSAUsername, "tsa-username", "", "TSA username")
	cmd.PersistentFlags().StringVar(&signatureTSAPassword, "tsa-password", "", "TSA password")
	cmd.PersistentFlags().StringVar(&certificateChainPath, "chain", "", "Certificate chain")
}

func parsePEMCertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&certificatePath, "crt", "", "Certificate path")
	cmd.PersistentFlags().StringVar(&privateKeyPath, "key", "", "Private key path")
}

func parsePKSC11CertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&pksc11LibPath, "lib", "", "Path to PKCS11 library")
	cmd.PersistentFlags().StringVar(&pksc11Pass, "pass", "", "PKCS11 password")
}

func parseInputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("in", "", "Input path")
	viper.BindPFlag("in", cmd.PersistentFlags().Lookup("in"))
}

func parseOutputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("out", "", "Output path")
	viper.BindPFlag("out", cmd.PersistentFlags().Lookup("out"))
}

func parseSignerName(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&signerNameFlag, "signer-name", "", "Signer name")
}

// Since viper is not supporting binding flags to an item of the array we use this workaround
func bindSignerFlagsToConfig(cmd *cobra.Command, c *signerConfig) {
	// SignData
	if cmd.PersistentFlags().Changed("approval") {
		c.SignData.Signature.Approval = signatureApproval
	}
	if cmd.PersistentFlags().Changed("type") {
		c.SignData.Signature.CertType = uint32(signatureType)
	}
	if cmd.PersistentFlags().Changed("info-name") {
		c.SignData.Signature.Info.Name = signatureInfoName
	}
	if cmd.PersistentFlags().Changed("info-location") {
		c.SignData.Signature.Info.Location = signatureInfoLocation
	}
	if cmd.PersistentFlags().Changed("info-reason") {
		c.SignData.Signature.Info.Reason = signatureInfoReason
	}
	if cmd.PersistentFlags().Changed("info-contact") {
		c.SignData.Signature.Info.ContactInfo = signatureInfoContact
	}
	if cmd.PersistentFlags().Changed("tsa-password") {
		c.SignData.TSA.URL = signatureTSAUrl
	}
	if cmd.PersistentFlags().Changed("tsa-url") {
		c.SignData.TSA.Password = signatureTSAPassword
	}

	// Certificate chain
	if cmd.PersistentFlags().Changed("chain") {
		c.CrtChainPath = certificateChainPath
	}

	// PEM
	if cmd.PersistentFlags().Changed("crt") {
		c.CrtPath = certificatePath
	}
	if cmd.PersistentFlags().Changed("key") {
		c.KeyPath = privateKeyPath
	}

	// PKSC11
	if cmd.PersistentFlags().Changed("lib") {
		c.LibPath = pksc11LibPath
	}
	if cmd.PersistentFlags().Changed("pass") {
		c.Pass = pksc11Pass
	}
}

func getChosenSignerConfig() signerConfig {
	var s signerConfig
	for _, s = range configSigners {
		if s.Name == signerNameFlag {
			return s
		}
	}
	log.Fatal(errors.New("signer not found"))
	return s
}
