package cmd

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
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

func parseSignDataSignatureFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("info-name", "", "Signature info name")
	cmd.PersistentFlags().String("info-location", "", "Signature info location")
	cmd.PersistentFlags().String("info-reason", "", "Signature info reason")
	cmd.PersistentFlags().String("info-contact", "", "Signature info contact")
	viper.BindPFlag("info.name", cmd.PersistentFlags().Lookup("info-name"))
	viper.BindPFlag("info.location", cmd.PersistentFlags().Lookup("info-location"))
	viper.BindPFlag("info.reason", cmd.PersistentFlags().Lookup("info-reason"))
	viper.BindPFlag("info.contact", cmd.PersistentFlags().Lookup("info-contact"))

	cmd.PersistentFlags().Bool("approval", false, "Approval")
	cmd.PersistentFlags().Uint("type", 0, "Certificate type")
	viper.BindPFlag("approval", cmd.PersistentFlags().Lookup("approval"))
	viper.BindPFlag("type", cmd.PersistentFlags().Lookup("type"))
}

func getSignDataSignature() sign.SignDataSignature {
	signDataSignature := sign.SignDataSignature{
		Info: sign.SignDataSignatureInfo{
			Name:        viper.GetString("info.name"),
			Location:    viper.GetString("info.location"),
			Reason:      viper.GetString("info.reason"),
			ContactInfo: viper.GetString("info.contact"),
			Date:        time.Now().Local(),
		},
		CertType: uint32(viper.GetInt("type")),
		Approval: viper.GetBool("approval"),
	}

	return signDataSignature
}

func parseSignDataTSAFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("tsa-url", "", "TSA url")
	cmd.PersistentFlags().String("tsa-username", "", "TSA username")
	cmd.PersistentFlags().String("tsa-password", "", "TSA password")
	viper.BindPFlag("tsa.url", cmd.PersistentFlags().Lookup("tsa-url"))
	viper.BindPFlag("tsa.username", cmd.PersistentFlags().Lookup("tsa-username"))
	viper.BindPFlag("tsa.password", cmd.PersistentFlags().Lookup("tsa-password"))
}

func getTSA() sign.TSA {
	tsa := sign.TSA{
		URL:      viper.GetString("tsa.url"),
		Username: viper.GetString("tsa.username"),
		Password: viper.GetString("tsa.password"),
	}
	return tsa
}

func parseCertificateChainPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("chain", "", "Certificate chain")
	viper.BindPFlag("chain", cmd.PersistentFlags().Lookup("chain"))
}

func parseOutputPathFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("out", "", "Output path")
	viper.BindPFlag("out", cmd.PersistentFlags().Lookup("out"))
}

func parseSSLCertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("crt", "", "Certificate path")
	cmd.PersistentFlags().String("key", "", "Private key path")
	viper.BindPFlag("crt", cmd.PersistentFlags().Lookup("crt"))
	viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key"))
}

func parsePKSC11CertificateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("lib", "", "Path to PKCS11 library")
	cmd.PersistentFlags().String("pass", "", "PKCS11 password")
	viper.BindPFlag("lib", cmd.PersistentFlags().Lookup("lib"))
	viper.BindPFlag("pass", cmd.PersistentFlags().Lookup("pass"))
}
