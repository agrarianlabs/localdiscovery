package discoverclient

// LookupLocalServiceIP look for the given service's ip
import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

// LookupLocalServiceIP look for the given service's ip
// in the discovery list.
// Expect to be a single line file containing the IPv4.
// The file name should be the service name.
func LookupLocalServiceIP(service, pth string) (string, error) {
	buf, err := ioutil.ReadFile(path.Join(pth, service))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("discovery file not present for %s", service)
		}
		return "", err
	}
	ipStr := strings.TrimSpace(string(buf))
	// validate the ip
	if _, _, err := net.ParseCIDR(strings.Split(ipStr, ":")[0] + "/32"); err != nil {
		return "", fmt.Errorf("invalid service ip (%s): %s", ipStr, err)
	}
	return ipStr, nil
}
