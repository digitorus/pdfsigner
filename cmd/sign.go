package cmd

import (
	"log"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign command",
	Long:  `Long multiline description here`,
}

var signPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Sign PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		signData := signer.NewSignData(viper.GetString("crt"), viper.GetString("key"), viper.GetString("chain"))
		signFilesByPatterns(filePatterns, signData)
	},
}

var signPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Signs PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		signData := signer.NewPKSC11SignData(viper.GetString("lib"), viper.GetString("pass"), viper.GetString("chain"))
		signFilesByPatterns(filePatterns, signData)
	},
}

func signFilesByPatterns(filePatterns []string, signData signer.SignData) {
	signData.TSA = getSignDataTSA()
	signData.Signature = getSignDataSignature()

	files, err := findFilesByPatterns(filePatterns)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if err := signer.SignFile(f, viper.GetString("out"), signData); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Signed PDF written to " + viper.GetString("out"))
}

func init() {
	RootCmd.AddCommand(signCmd)

	// Parse sign data flags
	parseSignDataSignatureFlags(signCmd)
	parseSignDataTSAFlags(signCmd)

	// Parse certificate chain path
	parseCertificateChainPathFlag(signCmd)

	// Parse output path
	parseOutputPathFlag(signCmd)

	//PEM sign command
	signCmd.AddCommand(signPEMCmd)
	parsePEMCertificateFlags(signPEMCmd)

	//PKSC11 sign command
	signCmd.AddCommand(signPKSC11Cmd)
	parsePKSC11CertificateFlags(signPKSC11Cmd)
}
