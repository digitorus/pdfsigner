package cmd

import (
	"log"
	"path/filepath"

	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch command",
	Long:  `Long multiline description here`,
}

var watchPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Watch PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}
		bindSignerFlagsToConfig(cmd, &c)
		c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		watch(c.SignData)
	},
}

var watchPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Watch PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := signerConfig{}
		bindSignerFlagsToConfig(cmd, &c)
		c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		watch(c.SignData)
	},
}

var watchBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with signer from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		c := getChosenSignerConfig()
		bindSignerFlagsToConfig(cmd, &c)

		switch c.Type {
		case "pem":
			c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		case "pksc11":
			c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		}

		watch(c.SignData)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	//PEM watch command
	watchCmd.AddCommand(watchPEMCmd)
	parsePEMCertificateFlags(watchPEMCmd)
	parseCommonFlags(watchPEMCmd)
	parseInputPathFlag(watchPEMCmd)
	parseOutputPathFlag(watchPEMCmd)

	//PKSC11 watch command
	watchCmd.AddCommand(watchPKSC11Cmd)
	parsePKSC11CertificateFlags(watchPKSC11Cmd)
	parseCommonFlags(watchPKSC11Cmd)
	parseOutputPathFlag(watchPKSC11Cmd)
	parseInputPathFlag(watchPEMCmd)

	// watch with watcher from config file
	watchCmd.AddCommand(watchBySignerNameCmd)
	parseSignerName(watchBySignerNameCmd)
	parseCommonFlags(watchBySignerNameCmd)
	parsePEMCertificateFlags(watchBySignerNameCmd)
	parsePKSC11CertificateFlags(watchBySignerNameCmd)
	parseInputPathFlag(watchBySignerNameCmd)
	parseOutputPathFlag(watchBySignerNameCmd)
}

func watch(signData signer.SignData) {
	watchFolder := viper.GetString("in")
	outputFolder := viper.GetString("out")
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
