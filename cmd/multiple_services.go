package cmd

import (
	"path"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsigner/files"
	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/queues/priority_queue"
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
		// fail if no config flag path provided
		requireConfig(cmd)

		// check if the config contains services
		if len(servicesConfig) < 1 {
			log.Fatal("no services found inside the config")
		}

		// setup wait group
		var wg sync.WaitGroup

		// setup services
		if len(serviceNames) > 1 {
			// setup services by name
			wg.Add(len(serviceNames))
			for _, n := range serviceNames {
				// get service config by name
				serviceConf := getConfigServiceByName(n)
				// setup service with signers
				setupServiceWithSigners(serviceConf, &wg)
			}
		} else {
			// setup all services
			wg.Add(len(servicesConfig))
			for _, s := range servicesConfig {
				// setup service with signers
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

// setupServiceWithSigners setup used by service signers and setup signer.
func setupServiceWithSigners(serviceConf serviceConfig, wg *sync.WaitGroup) {
	setupSigners(serviceConf.Type, serviceConf.Signer, serviceConf.Signers)

	go func(serviceConf serviceConfig) {
		setupService(serviceConf)
		wg.Done()
	}(serviceConf)
}

// directoryWatchersCount used to count the amount of directories watched. Required for license limits.
var directoryWatchersCount int

// setupSigners depending on the service type, watch or serve, setups the signer or signers
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

// setupSigner adds found inside the config by name signer to the queue for later use.
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
	signVerifyQueue.AddSignUnit(signerName, config.SignData)
}

// setupService depending on the type of the service setups service
func setupService(service serviceConfig) {
	if service.Type == "watch" {
		setupWatch(service.In, service.Out, service.Signer)
	} else if service.Type == "serve" {
		setupServe(service)
		log.Println("watch service", service.Name, "started")
	}
}

// setupWatch setups watcher which watches the input folder and adds the tasks to the queue.
func setupWatch(watchFolder, outputFilePath string, signerName string) {
	files.Watch(watchFolder, func(inputFilePath string) {

		// make signed file path
		_, fileName := path.Split(inputFilePath)
		fileNameArr := strings.Split(fileName, path.Ext(fileName))
		fileNameArr = fileNameArr[:len(fileNameArr)-1]
		fileNameNoExt := strings.Join(fileNameArr, "")
		signedFilePath := path.Join(outputFilePath, fileNameNoExt+"_signed"+path.Ext(fileName))

		// create session
		sessionID := signVerifyQueue.AddSignJob(signer.SignData{})

		// push job
		signVerifyQueue.AddTask(signerName, sessionID, inputFilePath, signedFilePath, priority_queue.LowPriority)
	})
}

// setupServe runs the web api according to the config settings
func setupServe(service serviceConfig) {
	// serve but only use allowed signers
	wa := webapi.NewWebAPI(service.Addr+":"+service.Port, signVerifyQueue, service.Signers, ver)
	wa.Serve()
}

// runQueues starts the mechanism to sign the files whenever they are getting into the queue.
func runQueues() {
	signVerifyQueue.Runner()
}

func init() {
	RootCmd.AddCommand(multiCmd)
}
