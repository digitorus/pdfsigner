package cmd

import (
	"log"
	"path"
	"strings"
	"sync"

	"bitbucket.org/digitorus/pdfsigner/files"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/priority_queue"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"bitbucket.org/digitorus/pdfsigner/webapi"
	"github.com/spf13/cobra"
)

// multiCmd represents the multi command
var multiCmd = &cobra.Command{
	Use:   "multiple-services",
	Short: "Run multiple services using the config file",
	Long:  `This command runs multiple services taken from the config file`,
	Run: func(cmd *cobra.Command, serviceNames []string) {
		requireConfig(cmd)

		if len(servicesConfig) < 1 {
			log.Fatal("no services found inside the config")
		}

		// setup wait group
		var wg sync.WaitGroup

		if len(serviceNames) > 1 {
			// setup services by name
			wg.Add(len(serviceNames))
			for _, n := range serviceNames {
				// get service config by name
				serviceConf := getConfigServiceByName(n)

				setupServiceWithSigners(serviceConf, &wg)
			}
		} else {
			// setup all services
			wg.Add(len(servicesConfig))
			for _, s := range servicesConfig {
				setupServiceWithSigners(s, &wg)
			}
		}

		// run queues
		runQueues()

		// run auto save license
		license.LD.AutoSave()

		// wait
		wg.Wait()
	},
}

func setupServiceWithSigners(serviceConf serviceConfig, wg *sync.WaitGroup) {
	setupSigners(serviceConf.Type, serviceConf.Signer, serviceConf.Signers)

	go func(serviceConf serviceConfig) {
		setupService(serviceConf)
		wg.Done()
	}(serviceConf)
}

var directoryWatchersCount int

func setupSigners(serviceType, configSignerName string, configSignerNames []string) {
	switch serviceType {
	case "watch":
		directoryWatchersCount++
		if directoryWatchersCount > license.LD.MaxDirectoryWatchers {
			log.Fatal("License: maximum directory watchers exceded, allowed:", license.LD.MaxDirectoryWatchers)
		}

		if len(configSignerNames) > 1 {
			log.Fatal(`Use signer instead of signers config setting for watch`)
		}

		if configSignerName != "" {
			setupSigner(configSignerName)
			return
		}
	case "serve":
		if configSignerName != "" {
			log.Fatal(`Use signers instead of signer config setting for serve`)
		}

		if len(configSignerNames) > 1 {
			for _, sn := range configSignerNames {
				setupSigner(sn)
			}
			return
		}
	default:
		log.Fatal("service type is not set inside the config")
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
		sessionID := qSign.AddJob(1, signer.SignData{})

		// push job
		qSign.AddTask(signerName, sessionID, inputFilePath, signedFilePath, priority_queue.LowPriority)
	})
}

func setupServe(service serviceConfig) {
	// serve but only use allowed signers
	wa := webapi.NewWebAPI(service.Addr+":"+service.Port, qSign, qVerify, service.Signers, ver)
	wa.Serve()
}

func runQueues() {
	qSign.Runner()
	qVerify.Runner()
}

func init() {
	RootCmd.AddCommand(multiCmd)
}
