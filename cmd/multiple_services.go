package cmd

import (
	"log"

	"bitbucket.org/digitorus/pdfsigner/files"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/webapi"
	"github.com/spf13/cobra"
)

// multiCmd represents the mixed command
var multiCmd = &cobra.Command{
	Use:   "mixed",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines`,
	Run: func(cmd *cobra.Command, serviceNames []string) {
		qSign = queued_sign.NewQSign()

		for _, n := range serviceNames {
			// get service by name
			service := getConfigServiceByName(n)
			setupSigners(service.Type, service.Signer, service.Signers)
			setupService(service)
		}
	},
}

func setupSigners(serviceType, configSignerName string, configSignerNames []string) {
	// only allow single signer string or array setting
	if configSignerName != "" && len(configSignerNames) > 1 {
		log.Fatal("please only use signer or signers setting for service")
	}

	if serviceType == "watch" && configSignerName != "" {
		setupSigner(configSignerName)
		return
	}

	if serviceType == "serve" && len(configSignerNames) > 1 {
		for _, sn := range configSignerNames {
			setupSigner(sn)
		}
		return
	}

	log.Fatal("no signers provided inside a config")

}

func setupSigner(signerName string) {
	// get config signer by name
	config := getSignerConfigByName(signerName)

	// set sign data
	switch config.Type {
	case "pem":
		config.SignData.SetPEM(config.CrtPath, config.KeyPath, config.CrtChainPath)
	case "pksc11":
		config.SignData.SetPKSC11(config.LibPath, config.Pass, config.CrtChainPath)
	}

	// add signer to signers map
	qSign.AddSigner(signerName, config.SignData, 10)
}

func setupService(service serviceConfig) {
	if service.Type == "watch" {
		setupWatch(service.In, service.Out, service.Signer)
	} else if service.Type == "serve" {
		setupServe(service)
	}
}

func setupWatch(watchFolder, outputFilePath string, signerName string) {
	files.Watch(watchFolder, func(inputFilePath string) {
		//c.SignData, service.In, service.Out
		outputFilePath := "path/to/output/file"

		qSign.PushJob("", signerName, inputFilePath, outputFilePath, priority_queue.LowPriority)
	})
}

func setupServe(service serviceConfig) {
	// serve but only use allowed signers
	wa := webapi.NewWebAPI(service.Addr, qSign, service.Signers)
	wa.Serve()
}

func init() {
	RootCmd.AddCommand(multiCmd)
}
