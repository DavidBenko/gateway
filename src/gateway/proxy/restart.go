package proxy

import (
	"gateway/config"
	"log"
	"strings"

	"gopkg.in/fsnotify.v1"
)

const (
	restartFile = "restart.txt"
)

func watchForRestarts(codePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Could not create filesystem watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(codePath)
	if err != nil {
		log.Fatalf("Could not watch code path '%s': %v",
			codePath, err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.HasSuffix(event.Name, restartFile) {
					log.Printf("%s Reloading code... TODO", config.Proxy)
				}
			case err := <-watcher.Errors:
				// TODO Handle?
				log.Println("error:", err)
			}
		}
	}()

	done := make(chan bool)
	<-done
}
