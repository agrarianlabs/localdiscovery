package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/agrarianlabs/localdiscovery"
	"github.com/creack/ehttp"
)

var (
	defaultPort      = 9090
	defaultDockerURL = "unix:///var/run/docker.sock"
)

// TODO: move this back to private repo with service controller.
func main() {
	var (
		listenAddr string
		dockerURL  string
	)

	if port := os.Getenv("PORT"); port == "" {
		listenAddr = fmt.Sprintf(":%d", defaultPort)
	} else {
		listenAddr = ":" + port
	}
	if url := os.Getenv("DOCKER_URL"); url == "" {
		dockerURL = defaultDockerURL
	} else {
		dockerURL = url
	}

	logrus.Printf("ready on %s", listenAddr)

	discovery, err := localdiscovery.NewDiscovery(dockerURL)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Fatal(http.ListenAndServe(listenAddr, ehttp.HandlerFunc(discovery.LookupHandler)))
}
