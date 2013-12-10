package rpc

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	defaultPortRangeMin uint16 = 10000
	defaultPortRangeMax uint16 = 11000
)

// Get port from an envirionment variable or fallback to the default.
func getPortFromEnvironment(env string, defaultPort uint16) uint16 {
	if port, err := strconv.ParseUint(os.Getenv(env), 10, 16); err == nil {
		return uint16(port)
	}
	return defaultPort
}

func NetListener() net.Listener {
	minPort := getPortFromEnvironment("PACKER_MIN_PORT", defaultPortRangeMin)
	maxPort := getPortFromEnvironment("PACKER_MAX_PORT", defaultPortRangeMax)
	return NetListenerInRange(minPort, maxPort)
}

// This finds an open port in the given range and returns a listener
// bound to that port.
func NetListenerInRange(min, max uint16) net.Listener {
	log.Printf("minimum port: %d\n", min)
	log.Printf("maximum port: %d\n", max)

	for port := min; port <= max; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		l, err := net.Listen("tcp", address)
		if err == nil {
			log.Printf("Listener address: %s\n", address)
			return l
		}
	}

	return nil
}
