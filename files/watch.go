package files

import (
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
)

type callback func(filePath string)

// Watch watches inside the provided folder and runs callback when event happened
func Watch(watchFolder string, cb callback) {
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
					inputFileExtension := strings.ToLower(path.Ext(inputFileName))

					if inputFileExtension == ".pdf" {
						cb(inputFileName)
					}
				}

			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()

	err = watcher.Add(watchFolder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
