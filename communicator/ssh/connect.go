package ssh

import (
	"errors"
	"log"
	"net"
	"time"
)

// ConnectFunc is a convenience method for returning a function
// that just uses net.Dial to communicate with the remote end that
// is suitable for use with the SSH communicator configuration.
func ConnectFunc(network, addr string, timeout time.Duration) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		timeoutCh := time.After(timeout)

		for {
			select {
			case <-timeoutCh:
				return nil, errors.New("timeout connecting to remote machine")
			default:
			}

			log.Printf("Opening conn for SSH to %s %s", network, addr)
			nc, err := net.DialTimeout(network, addr, 15*time.Second)
			if err == nil {
				return nc, nil
			}

			time.Sleep(500 * time.Millisecond)
		}
	}
}
