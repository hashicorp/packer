package rpc

import (
	"net"
	"net/rpc"
)

// rpcDial makes a TCP connection to a remote RPC server and returns
// the client. This will set the connection up properly so that keep-alives
// are set and so on and should be used to make all RPC connections within
// this package.
func rpcDial(address string) (*rpc.Client, error) {
	tcpConn, err := tcpDial(address)
	if err != nil {
		return nil, err
	}

	// Create an RPC client around our connection
	return rpc.NewClient(tcpConn), nil
}

// tcpDial connects via TCP to the designated address.
func tcpDial(address string) (*net.TCPConn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// Set a keep-alive so that the connection stays alive even when idle
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	return tcpConn, nil
}
