package rpc

import (
	"net"
	"testing"
)

func testClient(t *testing.T, server *Server) *Client {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		server.ServeConn(conn)
	}()

	clientConn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	client, err := NewClient(clientConn)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return client
}
