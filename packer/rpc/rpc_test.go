package rpc

import (
	"net"
	"net/rpc"
)

// This starts a RPC server for the given interface listening on the
// given address. The RPC server is ready when "readyChan" receives a message
// and the RPC server will quit when "stopChan" receives a message.
//
// This function should be run in a goroutine.
func testRPCServer(laddr string, name string, iface interface{}, readyChan chan int, stopChan <-chan int) {
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		panic(err)
	}

	// Close the listener when we exit so that the RPC server ends
	defer listener.Close()

	// Start the RPC server
	server := rpc.NewServer()
	server.RegisterName(name, iface)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// If there is an error, just ignore it.
				break
			}

			go server.ServeConn(conn)
		}
	}()

	// We're ready!
	readyChan <- 1

	// Block on waiting to receive from the channel
	<-stopChan
}
