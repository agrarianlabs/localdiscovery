package localdiscovery

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/agrarianlabs/localdiscovery/discoverclient"
)

// LookupHandler looks up the exposed port for a given container on the host.
// Method: POST
// Content-Type: application/json
// Request: (see HostInfo{})
//   - hostname (string): hostname of the target host
//   - ip       (string): ip of the target host
//   - mac      (string): hardware address of the target host
//   - port     (string): port as a string. ex: 80, 8080/tcp, 8125/udp
// Response:
//   - port        (int): first exposed port public value. 0 means not exposed.
func (d *DockerDiscovery) LookupHandler(w http.ResponseWriter, req *http.Request) error {
	lookupReq := discoverclient.LookupRequest{}
	err := json.NewDecoder(req.Body).Decode(&lookupReq)
	_ = req.Body.Close() // best effort.
	if err != nil {
		return err
	}
	port, err := d.LookupPort(strings.Split(req.RemoteAddr, ":")[0], lookupReq.Port)
	if err != nil {
		return err
	}
	logrus.Printf("Lookup result for %s:%s is %d", req.RemoteAddr, lookupReq.Port, port)
	return json.NewEncoder(w).Encode(port)
}
