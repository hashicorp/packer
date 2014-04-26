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

	// The server side
	go func() {
		defer close(doneCh)

		s0, err := server.Accept(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		s1, err := server.Dial(1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			defer s1.Close()
			data := readStream(t, s1)
			if data != "another" {
				t.Fatalf("bad: %#v", data)
			}
		}()

		go func() {
			defer wg.Done()
			defer s0.Close()
			data := readStream(t, s0)
			if data != "hello" {
				t.Fatalf("bad: %#v", data)
			}
		}()

		wg.Wait()
	}()

	s0, err := client.Dial(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	s1, err := client.Accept(1)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := s0.Write([]byte("hello")); err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := s1.Write([]byte("another")); err != nil {
		t.Fatalf("err: %s", err)
	}

	s0.Close()
	s1.Close()

	// Wait for the server to be done
	<-doneCh
}

func TestMuxConn_lotsOfData(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	// When the server is done
	doneCh := make(chan struct{})

	// The server side
	go func() {
		defer close(doneCh)

		s0, err := server.Accept(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		var data [1024]byte
		for {
			n, err := s0.Read(data[:])
			if err == io.EOF {
				break
			}

			dataString := string(data[0:n])
			if dataString != "hello" {
				t.Fatalf("bad: %#v", dataString)
			}
		}

		s0.Close()
	}()

	s0, err := client.Dial(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	for i := 0; i < 4096*4; i++ {
		if _, err := s0.Write([]byte("hello")); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	if err := s0.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Wait for the server to be done
	<-doneCh
}

// This tests that even when the client end is closed, data can be
// read from the server.
func TestMuxConn_clientCloseRead(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	// This channel will be closed when we close
	waitCh := make(chan struct{})

	go func() {
		conn, err := server.Accept(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		<-waitCh

		_, err = conn.Write([]byte("foo"))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		conn.Close()
	}()

	s0, err := client.Dial(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := s0.Close(); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Close this to continue on on the server-side
	close(waitCh)

	var data [1024]byte
	n, err := s0.Read(data[:])
	if string(data[:n]) != "foo" {
		t.Fatalf("bad: %#v", string(data[:n]))
	}
}

func TestMuxConn_socketClose(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	go func() {
		_, err := server.Accept(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		server.rwc.Close()
	}()

	s0, err := client.Dial(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var data [1024]byte
	_, err = s0.Read(data[:])
	if err != io.EOF {
		t.Fatalf("err: %s", err)
	}
}

func TestMuxConn_clientClosesStreams(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	go func() {
		conn, err := server.Accept(0)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		conn.Close()
	}()

	s0, err := client.Dial(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

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
	go server.Accept(0)

	s0, err := client.Dial(0)
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

func TestMuxConnNextId(t *testing.T) {
	client, server := testMux(t)
	defer client.Close()
	defer server.Close()

	a := client.NextId()
	b := client.NextId()

	if a != 1 || b != 2 {
		t.Fatalf("IDs should increment")
	}

	a = server.NextId()
	b = server.NextId()

	if a != 1 || b != 2 {
		t.Fatalf("IDs should increment: %d %d", a, b)
	}
}
