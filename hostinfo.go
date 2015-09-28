package discover

import (
	"fmt"
	"net"
	"os"
)

// HostInfo represent basic host info.
type HostInfo struct {
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	MacAddress string `json:"mac"`
}

// Common errors.
var (
	ErrNoAddressFound = fmt.Errorf("no address found for the given interface")
)

// LookupHostInfo tries to retrieve the hostname and the ip&mac address of the given interface.
func LookupHostInfo(ifaceName string) (HostInfo, error) {
	ret := HostInfo{}

	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return ret, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ret, err
	}
	if len(addrs) == 0 {
		return ret, ErrNoAddressFound
	}

	var ipv4 string
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return ret, err
		}
		if ip.To4() != nil {
			ipv4 = ip.String()
			break
		}
	}
	if ipv4 == "" {
		return ret, ErrNoAddressFound
	}
	return HostInfo{
		Hostname:   hostname,
		IP:         ipv4,
		MacAddress: iface.HardwareAddr.String(),
	}, nil
}
