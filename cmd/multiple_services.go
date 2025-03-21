package cmd

import (
	"fmt"
	"slices"
	"sync"

	"github.com/digitorus/pdfsigner/files"
	"github.com/digitorus/pdfsigner/license"
	"github.com/digitorus/pdfsigner/queues/priority_queue"
	"github.com/digitorus/pdfsigner/queues/queue"
	"github.com/digitorus/pdfsigner/webapi"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// multiCmd represents the multi command.
var multiCmd = &cobra.Command{
	Use:   "services",
	Short: "Run multiple services using the config file",
	RunE: func(cmd *cobra.Command, serviceNames []string) error {
		// loading jobs from the db
		err := signVerifyQueue.LoadFromDB()
		if err != nil {
			return err
		}

		// check if the config contains services
		if len(config.Services) < 1 {
			return fmt.Errorf("no services found inside the config")
		}

		// setup wait group
		var wg sync.WaitGroup

		// setup services
		for sname, sconfig := range config.Services {
			// setup services by name
			if len(serviceNames) == 0 {
				wg.Add(len(sname))
				setupServiceWithSigners(sconfig, &wg)
			} else if slices.Contains(serviceNames, sname) {
				wg.Add(len(sname))
				setupServiceWithSigners(sconfig, &wg)
			}
		}

		// run queues
		runQueues()

		// run auto save license
		license.LD.AutoSave()

		// wait
		wg.Wait()
		return nil
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

// setupSigners depending on the service type, watch or serve, setups the signer or signers.
func setupSigners(serviceType, configSignerName string, configSignerNames []string) {
	switch serviceType {
	case "watch":
		directoryWatchersCount++
		if directoryWatchersCount > license.LD.MaxDirectoryWatchers {
			log.Fatal("License: maximum directory watchers exceded, allowed:", license.LD.MaxDirectoryWatchers)
		}

		// check if array of signer names were provided watch service may only contain single signer
		if len(configSignerNames) > 1 {
			log.Fatal(`Use signer instead of signers config setting for watch`)
		}

		// setup signer
		if configSignerName != "" {
			err := setupSigner(configSignerName)
			if err != nil {
				log.Fatalf("failed to setup signer: %s", err)
			}

			return
		}
	case "serve":
		// serve service should contain array of signers instead of single signer
		if configSignerName != "" {
			log.Fatal(`Use signers instead of signer config setting for serve`)
		}

		// setup signers
		if len(configSignerNames) > 1 {
			for _, sn := range configSignerNames {
				err := setupSigner(sn)
				if err != nil {
					log.Fatalf("failed to setup signer: %s", err)
				}
			}
		}

		// setup verifier unit
		setupVerifier()
	default:
		log.Fatal("service type is not set inside the config")
	}
}

// setupSigner adds found inside the config by name signer to the queue for later use.
func setupSigner(signerName string) error {
	// get config signer by name
	config, err := getSignerConfigByName(signerName)
	if err != nil {
		return err
	}

	switch config.Type {
	case "pem":
		err = config.SignData.SetPEM(config.Cert, config.Key, config.Chain)
	case "pksc11":
		err = config.SignData.SetPKSC11(config.Lib, config.Pass, config.Chain)
	}
	if err != nil {
		return err
	}

	// add signer to signers map
	signVerifyQueue.AddSignUnit(signerName, config.SignData)

	return nil
}

// setupService depending on the type of the service setups service.
func setupService(service serviceConfig) {
	var err error
	if service.Type == "watch" {
		err = setupWatch(service)
	} else if service.Type == "serve" {
		err = setupServe(service)
	}

	if err != nil {
		log.Fatalf("failed to setup %s: %s", service.Type, err)
	}
}

// setupWatch setups watcher which watches the input folder and adds the tasks to the queue.
func setupWatch(service serviceConfig) error {
	files.Watch(service.In, func(inputFilePath string, left int) {
		// make signed file path
		signedFilePath := getOutputFilePathByInputFilePath(inputFilePath, service.Out)

		// create session
		jobID := signVerifyQueue.AddSignJob(queue.JobSignConfig{
			ValidateSignature: service.ValidateSignature,
		})

		// push job to the queue
		_, err := signVerifyQueue.AddTask(service.Signer, jobID, "", inputFilePath, signedFilePath, priority_queue.LowPriority)
		if err != nil {
			log.Debugf("failed to add task: %s", err)
		}
		if left == 0 {
			err = signVerifyQueue.SaveToDB(jobID)
			if err != nil {
				log.Debugf("failed to save job to db: %s", err)
			}
		}
	})

	return nil
}

// setupServe runs the web api according to the config settings.
func setupServe(service serviceConfig) error {
	// serve but only use allowed signers
	wa := webapi.NewWebAPI(service.Addr+":"+service.Port, signVerifyQueue, service.Signers, ver, service.ValidateSignature)
	wa.Serve()

	return nil
}

// runQueues starts the mechanism to sign the files whenever they are getting into the queue.
func runQueues() {
	signVerifyQueue.StartProcessor()
}

func init() {
	RootCmd.AddCommand(multiCmd)
	// No need to call parseConfigFlag since config is now a global flag
}
