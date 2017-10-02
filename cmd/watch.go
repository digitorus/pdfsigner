package cmd

import (
	"log"

	"path/filepath"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	watchInputPath  string
	watchOutputPath string
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long:  `Long multiline description`,
}

var watchSSLCmd = &cobra.Command{
	Use:   "ssl",
	Short: "Watch and sign PDF with SSL",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		signData := signer.NewSignData(viper.GetString("crt"), viper.GetString("key"), viper.GetString("chain"))
		watch(signData)
	},
}

var watchPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Watch and sign PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		signData := signer.NewPKSC11SignData(viper.GetString("lib"), viper.GetString("pass"), viper.GetString("chain"))
		watch(signData)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
	// Parse sign data flags
	parseSignDataSignatureFlags(watchCmd)
	parseSignDataTSAFlags(watchCmd)

	// Parse certificate chain path
	parseCertificateChainPathFlag(watchCmd)

	//SSL watch command
	watchCmd.AddCommand(watchSSLCmd)
	parseSSLCertificateFlags(watchSSLCmd)

	//PKSC11 watch command
	watchCmd.AddCommand(watchPKSC11Cmd)
	parsePKSC11CertificateFlags(watchPKSC11Cmd)

	watchCmd.PersistentFlags().StringVar(&watchInputPath, "in", "", "Watch input folder path")
	watchCmd.PersistentFlags().StringVar(&watchOutputPath, "out", "", "Output folder path")
}

func watch(signData signer.SignData) {
	signData.TSA = getTSA()
	signData.Signature = getSignDataSignature()

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
						fullFilePath := filepath.Join(watchInputPath, inputFileName)
						signer.SignFile(fullFilePath, watchOutputPath, signData)
					}
					log.Println("created file:", event.Name)

				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watchInputPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
