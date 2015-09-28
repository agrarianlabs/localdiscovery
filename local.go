package discover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-fsnotify/fsnotify"
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

// LookupLocalServiceIP look for the given service's ip
// in the discovery list.
// Expect to be a single line file containing the IPv4.
// The file name should be the service name.
func LookupLocalServiceIP(service, pth string) (string, error) {
	buf, err := ioutil.ReadFile(path.Join(pth, service))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("discovery file not present")
		}
		return "", err
	}
	ipStr := strings.TrimSpace(string(buf))
	// validate the ip
	if _, _, err := net.ParseCIDR(ipStr + "/32"); err != nil {
		return "", fmt.Errorf("invalid service ip (%s): %s", ipStr, err)
	}
	return ipStr, nil
}

// LocalLookup looks up the publicly exposed port for the current host.
// First lookup the local host infos, then sends the port lookup request.
// - url is tha address of the discover service.
// - iface is the network interface to lookup.
// - port is a string and can contain /udp or /tcp suffix.
func LocalLookup(url, iface, port string) (int, error) {
	hostInfo, err := LookupHostInfo(iface)
	if err != nil {
		return -1, err
	}
	buf, err := json.Marshal(LookupRequest{HostInfo: hostInfo, Port: port})
	if err != nil {
		return -1, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return -1, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	buf, err = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close() // best effort.
	if err != nil {
		return -1, err
	}
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected response: %d (%s)", resp.StatusCode, buf)
	}
	var exposedPort int
	if err := json.Unmarshal(buf, &exposedPort); err != nil {
		return -1, err
	}
	return exposedPort, nil
}
