package localdiscovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

// Discovery holds the state.
type Discovery struct {
	client *docker.Client
}

// NewDiscovery instantiates a new Discovery object.
func NewDiscovery(dockerAddr string) (*Discovery, error) {
	client, err := docker.NewClient(dockerAddr)
	if err != nil {
		return nil, err
	}
	return &Discovery{
		client: client,
	}, nil
}

// LookupPort lookup the given port for a container and return the first port value.
// First tries to lookup the container with hostname as ID, then lookup the hostname.
func (d *Discovery) LookupPort(hostname, ip, mac, port string) (int, error) {
	// default to TCP if not specified.
	if strings.Index(port, "/") == -1 {
		port += "/tcp"
	}
	log := logrus.WithFields(logrus.Fields{"lookup_hostname": hostname, "port": port})
	// small return helper.
	returnPort := func(cont *docker.Container) (int, error) {
		ports, ok := cont.NetworkSettings.Ports[docker.Port(port)]
		if !ok {
			log.Warning("The port is not exposed")
			return -1, nil
		}
		port, err := strconv.Atoi(strings.Split(ports[0].HostPort, "/")[0])
		if err != nil {
			log.WithError(err).Error("Invalid port format")
			return -1, err
		}
		return port, nil
	}

	if cont, err := d.client.InspectContainer(hostname); err != nil {
		// we have an error, if it is different thant NoSuchContainer,
		// terminates. Otherwise, lookup ip.
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			return -1, err
		}
	} else {
		// We found the container, return the ports.
		return returnPort(cont)
	}

	// Lookup all containers.
	containers, err := d.client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return -1, err
	}
	for _, cont := range containers {
		// Fetch more details about that container.
		cont, err := d.client.InspectContainer(cont.ID)
		if err != nil {
			log.WithError(err).Error("error inspecting container, skipping")
			continue
		}
		// If the hostname and IP match, we return the port.
		if cont.Config.Hostname == hostname && cont.NetworkSettings.IPAddress == ip && cont.NetworkSettings.MacAddress == mac {
			return returnPort(cont)
		}

	}
	// If we reach this point, we didn't find the container.
	return -1, fmt.Errorf("unable to lookup the port %s for container %s (%s)", port, hostname, ip)
}
