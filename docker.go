package localdiscovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

// DockerDiscovery creates a http service
// to lookup the exposed port of a given container.
type DockerDiscovery struct {
	client *docker.Client
}

// NewDockerDiscovery instantiates a new DockerDiscovery object.
// Connects to docker and store the socket.
func NewDockerDiscovery(dockerAddr string) (*DockerDiscovery, error) {
	client, err := docker.NewClient(dockerAddr)
	if err != nil {
		return nil, err
	}
	return &DockerDiscovery{
		client: client,
	}, nil
}

// LookupPort lookup the given port for a container and return the first port value.
// It matches based on the IP of the caller.
func (d *DockerDiscovery) LookupPort(ip, port string) (int, error) {
	// default to TCP if not specified.
	if strings.Index(port, "/") == -1 {
		port += "/tcp"
	}
	log := logrus.WithField("port", port).WithField("ip", ip)
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
		if cont.NetworkSettings.IPAddress == ip {
			return returnPort(cont)
		}

	}
	// If we reach this point, we didn't find the container.
	return -1, fmt.Errorf("unable to lookup the port %s for container %s", port, ip)
}
