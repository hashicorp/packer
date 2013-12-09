package rpc

import (
	"io"
	"net"
	"sync"
	"testing"
)

func readStream(t *testing.T, s io.Reader) string {
	var data [1024]byte
	n, err := s.Read(data[:])
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return string(data[0:n])
}

func testMux(t *testing.T) (client *MuxConn, server *MuxConn) {
	l, err := net.Listen("tcp", ":0")
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

		server = NewMuxConn(conn)
	}()

	// Client side
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	client = NewMuxConn(conn)

	// Wait for the server
	<-doneCh

	return
}

func TestMuxConn(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	// When the server is done
	doneCh := make(chan struct{})
	readyCh := make(chan struct{})

	// The server side
	go func() {
		defer close(doneCh)

		s0, err := server.Stream(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		s1, err := server.Stream(1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		close(readyCh)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			data := readStream(t, s1)
			if data != "another" {
				t.Fatalf("bad: %#v", data)
			}
		}()

		go func() {
			defer wg.Done()
			data := readStream(t, s0)
			if data != "hello" {
				t.Fatalf("bad: %#v", data)
			}
		}()

		wg.Wait()
	}()

	s0, err := client.Stream(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	s1, err := client.Stream(1)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Wait for the server to be ready
	<-readyCh

	if _, err := s0.Write([]byte("hello")); err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := s1.Write([]byte("another")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Wait for the server to be done
	<-doneCh
}

func TestMuxConn_clientClosesStreams(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	s0, err := client.Stream(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// This should block forever since we never write onto this stream.
	var data [1024]byte
	_, err = s0.Read(data[:])
	if err != io.EOF {
		t.Fatalf("err: %s", err)
	}
}

func TestMuxConn_serverClosesStreams(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	s0, err := client.Stream(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// This should block forever since we never write onto this stream.
	var data [1024]byte
	_, err = s0.Read(data[:])
	if err != io.EOF {
		t.Fatalf("err: %s", err)
	}
}
