package localdiscovery

import (
	"log"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	fsnotify "github.com/go-fsnotify/fsnotify"
)

// WatchService start a watcher on the given service.
// Execute preHook right away and after each event on the target service.
// Execute postHook only after the event.
func WatchService(preHook, postHook func(ip string), service, discoveryPath string, stopChan <-chan struct{}) {
	if discoveryPath == "" || service == "" || preHook == nil || stopChan == nil {
		logrus.Errorf("WatchService fail for %q: preHook, service, discoveryPath and stopChan are mandatory", service)
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }() // Best effort.

	watcher.Add(discoveryPath)
	path := path.Join(discoveryPath, service)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		ip, err := LookupLocalServiceIP(service, discoveryPath)
		if err != nil {
			// TODO:  handle error.
		}
		preHook(ip)

	wait:
		/// Wait for an event
		select {
		case <-ticker.C:
		case <-stopChan:
			return
		case event, open := <-watcher.Events:
			if !open {
				return
			}
			if event.Name != path {
				goto wait
			}
		case err, open := <-watcher.Errors:
			if !open {
				return
			}
			// TODO: handle errors.
			_ = err
		}

		// TODO: create a post hook? (for Close())
		if postHook != nil {
			postHook(ip)
		}
	}
}
