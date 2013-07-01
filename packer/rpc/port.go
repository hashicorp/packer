package rpc

import (
	"fmt"
	"net"
)

var portRangeMin int = 10000
var portRangeMax int = 11000

// This sets the port range that the RPC stuff will use when creating
// new temporary servers. Some RPC calls require the creation of temporary
// RPC servers. These allow you to pick a range these bind to.
func PortRange(min, max int) {
	portRangeMin = min
	portRangeMax = max
}

// This finds an open port in the given range and returns a listener
// bound to that port.
func netListenerInRange(min, max int) net.Listener {
	for port := min; port <= max; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return l
		}
	}

	return nil
}
