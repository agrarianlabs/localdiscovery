package discoverclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// LookupRequest is the data send via POST for the Lookup Handler.
type LookupRequest struct {
	Port string
}

// SelfDockerLookup looks up the publicly exposed port for the current host.
// First lookup the local host infos, then sends the port lookup request.
// - url is the address of the discover service.
// - iface is the network interface to lookup.
// - port is a string and may contain /udp or /tcp suffix.
func SelfDockerLookup(url, iface, port string) (int, error) {
	buf, err := json.Marshal(LookupRequest{Port: port})
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
