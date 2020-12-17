package ssh

import (
	"io"
	"log"
	"net"
)

// ProxyServe starts Accepting connections
func ProxyServe(l net.Listener, done <-chan struct{}, dialer func() (net.Conn, error)) {
	for {
		// Accept will return if either the underlying connection is closed or if a connection is made.
		// after returning, check to see if c.done can be received. If so, then Accept() returned because
		// the connection has been closed.
		client, err := l.Accept()
		select {
		case <-done:
			log.Printf("[WARN] Tunnel: received Done event: %v", err)
			return
		default:
			if err != nil {
				log.Printf("[ERROR] Tunnel: listen.Accept failed: %v", err)
				continue
			}
			log.Printf("[DEBUG] Tunnel: client '%s' accepted", client.RemoteAddr())
			// Proxy bytes from one side to the other
			go handleProxyClient(client, dialer)
		}
	}
}

// handleProxyClient will open a connection using the dialer, and ensure close events propagate to the brokers
func handleProxyClient(clientConn net.Conn, dialer func() (net.Conn, error)) {
	//We have a client connected, open an upstream connection to the destination
	upstreamConn, err := dialer()
	if err != nil {
		log.Printf("[ERROR] Tunnel: failed to open connection to upstream: %v", err)
		clientConn.Close()
		return
	}

	// channels to wait on the close event for each connection
	serverClosed := make(chan struct{}, 1)
	upstreamClosed := make(chan struct{}, 1)

	go brokerData(clientConn, upstreamConn, upstreamClosed)
	go brokerData(upstreamConn, clientConn, serverClosed)

	// Now we wait for the connections to close and notify the other side of the event
	select {
	case <-upstreamClosed:
		clientConn.Close()
		<-serverClosed
	case <-serverClosed:
		upstreamConn.Close()
		<-upstreamClosed
	}
	log.Printf("[DEBUG] Tunnel: client ('%s') proxy closed", clientConn.RemoteAddr())
}

// brokerData is responsible for copying data src => dest. It will also close the src when there are no more bytes to transfer
func brokerData(src net.Conn, dest net.Conn, srcClosed chan struct{}) {
	_, err := io.Copy(src, dest)
	if err != nil {
		log.Printf("[ERROR] Tunnel: Copy error: %s", err)
	}
	if err := src.Close(); err != nil {
		log.Printf("[ERROR] Tunnel: Close error: %s", err)
	}
	srcClosed <- struct{}{}
}
