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

func TestMuxConn(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// When the server is done
	doneCh := make(chan struct{})
	readyCh := make(chan struct{})

	// The server side
	go func() {
		defer close(doneCh)
		conn, err := l.Accept()
		l.Close()
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		defer conn.Close()

		mux := NewMuxConn(conn)
		s0, err := mux.Stream(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		s1, err := mux.Stream(1)
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

	// Client side
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer conn.Close()

	mux := NewMuxConn(conn)
	s0, err := mux.Stream(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	s1, err := mux.Stream(1)
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
