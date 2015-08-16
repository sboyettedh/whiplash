package whiplash // github.com/sboyettedh/whiplash

import (
	"fmt"
	"net"
	"regexp"
)

// getIface tests to see if a given address (or fractional address,
// IPv4 or IPv6) is attached to an interface on a machine. This is
// useful if you have many machines, each with multiple interfaces,
// each interface on a different network, and you wish a network
// service to listen only on a given network.
//
// If a machine has address 10.2.1.99 on an interface, then
// `getIface("10.1.2")` will return a net.Addr. If no interface has an
// IP matching that string, or if there are multiple addresses
// matching it, an error will be returned.
func getIface(addr string) (net.Addr, error) {
	ias, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var matches int
	var iaddr net.Addr
	for _, ia := range ias {
		matched, _ := regexp.MatchString(addr, ia.String())
		if matched == true {
			iaddr = ia
			matches++
		}
	}
	if matches == 0 {
		return nil, fmt.Errorf("No interfaces match '%v'", addr)
	}
	if matches > 1 {
		return nil, fmt.Errorf("Multiple interfaces match '%v'", addr)
	}
	return iaddr, nil
}
