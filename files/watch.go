package files

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type callback func(filePath string)

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
					inputFileExtension := filepath.Ext(inputFileName)

					if inputFileExtension == "pdf" {
						fullFilePath := filepath.Join(watchFolder, inputFileName)

						cb(fullFilePath)
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
