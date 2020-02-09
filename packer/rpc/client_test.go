package rpc

import (
	"net"
	"testing"
)

func testConn(t *testing.T) (net.Conn, net.Conn) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var serverConn net.Conn
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		defer l.Close()
		var err error
		serverConn, err = l.Accept()
		if err != nil {
			errChan <- err
			return
		}
	}()

	clientConn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = <-errChan
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return clientConn, serverConn
}

func testClientServer(t *testing.T) (*Client, *Server) {
	clientConn, serverConn := testConn(t)

	server, err := NewServer(serverConn)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go server.Serve()

	client, err := NewClient(clientConn)
	if err != nil {
		server.Close()
		t.Fatalf("err: %s", err)
	}

	return client, server
}
