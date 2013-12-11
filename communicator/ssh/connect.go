package ssh

import (
	"net"
	"time"
)

// ConnectFunc is a convenience method for returning a function
// that just uses net.Dial to communicate with the remote end that
// is suitable for use with the SSH communicator configuration.
func ConnectFunc(network, addr string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		return net.DialTimeout(network, addr, 15*time.Second)
	}
}
