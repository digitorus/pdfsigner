package files

import (
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type callback func(filePath string, left int)

// Watch watches inside the provided folder and runs callback when event happened.
func Watch(watchFolder string, cb callback) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					inputFileName := event.Name
					inputFileExtension := strings.ToLower(path.Ext(inputFileName))

					if inputFileExtension == ".pdf" {
						cb(inputFileName, len(watcher.Events))
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
