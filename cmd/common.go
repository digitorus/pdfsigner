package cmd

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// common flags
	signerNameFlag           string
	certificateChainPathFlag string
	inputFileNameFlag        string

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

func getConfigSignerByName(signerName string) signerConfig {
	var s signerConfig
	for _, s = range signersConfig {
		if s.Name == signerName {
			return s
		}
	}
	log.Fatal(errors.New("signer not found"))
	return s
}

func getConfigServiceByName(serviceName string) serviceConfig {
	var s serviceConfig
	for _, s = range servicesConfig {
		if s.Name == serviceName {
			return s
		}
	}
	log.Fatal(errors.New("service not found"))
	return s
}

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

func watch(signData signer.SignData, watchFolder, outputFolder string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					inputFileName := event.Name
					inputFileExtension := filepath.Ext(inputFileName)
					if inputFileExtension == "pdf" {
						fullFilePath := filepath.Join(watchFolder, inputFileName)
						signer.SignFile(fullFilePath, outputFolder, signData)
					}
					log.Println("created file:", event.Name)

				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watchFolder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func signFilesByPatterns(filePatterns []string, signData signer.SignData) {
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