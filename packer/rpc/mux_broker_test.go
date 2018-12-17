package rpc

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/yamux"
)

func TestMuxBroker(t *testing.T) {
	c, s := testYamux(t)
	defer c.Close()
	defer s.Close()

	bc := newMuxBroker(c)
	bs := newMuxBroker(s)
	go bc.Run()
	go bs.Run()

	errChan := make(chan error, 2)
	go func() {
		defer close(errChan)
		c, err := bc.Dial(5)
		if err != nil {
			errChan <- fmt.Errorf("err dialing: %s", err.Error())
			return
		}

		if _, err := c.Write([]byte{42}); err != nil {
			errChan <- fmt.Errorf("err writing: %s", err.Error())
		}
	}()

	client, err := bs.Accept(5)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var data [1]byte
	if _, err := client.Read(data[:]); err != nil {
		t.Fatalf("err: %s", err)
	}

	if data[0] != 42 {
		t.Fatalf("bad: %d", data[0])
	}

	for err := range errChan {
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func testYamux(t *testing.T) (client *yamux.Session, server *yamux.Session) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Server side
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		conn, err := l.Accept()
		l.Close()
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		server, err = yamux.Server(conn, nil)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}()

	// Client side
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	client, err = yamux.Client(conn, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Wait for the server
	<-doneCh

	return
}
