package cmd

import (
	"log"
	"path"
	"strings"
	"sync"

	"bitbucket.org/digitorus/pdfsigner/files"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"bitbucket.org/digitorus/pdfsigner/webapi"
	"github.com/spf13/cobra"
)

// multiCmd represents the mixed command
var multiCmd = &cobra.Command{
	Use:   "multiple-services",
	Short: "Runs multiple services of the config file",
	Long:  `This command runs multiple services taken from the config file`,
	Run: func(cmd *cobra.Command, serviceNames []string) {
		requireConfig(cmd)

		if len(serviceNames) < 1 {
			log.Fatal("no service names provided")
		}

		var wg sync.WaitGroup
		wg.Add(len(serviceNames))

		for _, n := range serviceNames {
			// get service config by name
			serviceConf := getConfigServiceByName(n)
			setupSigners(serviceConf.Type, serviceConf.Signer, serviceConf.Signers)

			go func(serviceConf serviceConfig) {
				setupService(serviceConf)
				wg.Done()
			}(serviceConf)
		}

		runQueues()

		wg.Wait()
	},
}

func setupSigners(serviceType, configSignerName string, configSignerNames []string) {
	if serviceType == "watch" {
		if len(configSignerNames) > 1 {
			log.Fatal(`Use signer instead of signers config setting for watch`)
		}

		if configSignerName != "" {
			setupSigner(configSignerName)
			return
		}
	}

	if serviceType == "serve" {
		if configSignerName != "" {
			log.Fatal(`Use signers instead of signer config setting for serve`)
		}

		if len(configSignerNames) > 1 {
			for _, sn := range configSignerNames {
				setupSigner(sn)
			}
			return
		}
	}
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
		log.Println("watch service", service.Name, "started")
	}
}

func setupWatch(watchFolder, outputFilePath string, signerName string) {
	files.Watch(watchFolder, func(inputFilePath string) {

		// make signed file path
		_, fileName := path.Split(inputFilePath)
		fileNameArr := strings.Split(fileName, path.Ext(fileName))
		fileNameArr = fileNameArr[:len(fileNameArr)-1]
		fileNameNoExt := strings.Join(fileNameArr, "")
		signedFilePath := path.Join(outputFilePath, fileNameNoExt+"_signed"+path.Ext(fileName))

		// create session
		sessionID := qSign.NewSession(1, signer.SignData{})

		// push job
		qSign.PushJob(signerName, sessionID, inputFilePath, signedFilePath, priority_queue.LowPriority)
	})
}

func setupServe(service serviceConfig) {
	// serve but only use allowed signers
	wa := webapi.NewWebAPI(service.Addr+":"+service.Port, qSign, qVerify, service.Signers)
	wa.Serve()
}

func runQueues() {
	qSign.Runner()
	qVerify.Runner()
}

func init() {
	RootCmd.AddCommand(multiCmd)
}
