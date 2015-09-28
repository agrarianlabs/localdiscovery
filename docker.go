package localdiscovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
// First tries to lookup the container with hostname as ID, then lookup the hostname.
func (d *DockerDiscovery) LookupPort(hostname, ip, mac, port string) (int, error) {
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

// SelfDockerLookup looks up the publicly exposed port for the current host.
// First lookup the local host infos, then sends the port lookup request.
// - url is tha address of the discover service.
// - iface is the network interface to lookup.
// - port is a string and can contain /udp or /tcp suffix.
func SelfDockerLookup(url, iface, port string) (int, error) {
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
