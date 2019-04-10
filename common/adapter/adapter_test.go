package adapter

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"

	"golang.org/x/crypto/ssh"
)

func TestAdapter_Serve(t *testing.T) {

	// done signals the adapter that the provisioner is done
	done := make(chan struct{})

	acceptC := make(chan struct{})
	l := listener{done: make(chan struct{}), acceptC: acceptC}

	config := &ssh.ServerConfig{}

	ui := new(packer.NoopUi)

	sut := NewAdapter(done, &l, config, "", ui, communicator{})
	go func() {
		i := 0
		for range acceptC {
			i++
			if i == 4 {
				close(done)
				l.Close()
			}
		}
	}()

	sut.Serve()
}

type listener struct {
	done    chan struct{}
	acceptC chan<- struct{}
	i       int
}

func (l *listener) Accept() (net.Conn, error) {
	log.Println("Accept() called")
	l.acceptC <- struct{}{}
	select {
	case <-l.done:
		log.Println("done, serving an error")
		return nil, errors.New("listener is closed")

	case <-time.After(10 * time.Millisecond):
		l.i++

		if l.i%2 == 0 {
			c1, c2 := net.Pipe()

			go func(c net.Conn) {
				<-time.After(100 * time.Millisecond)
				log.Println("closing c")
				c.Close()
			}(c1)

			return c2, nil
		}
	}

	return nil, errors.New("accept error")
}

func (l *listener) Close() error {
	close(l.done)
	return nil
}

func (l *listener) Addr() net.Addr {
	return addr{}
}

type addr struct{}

func (a addr) Network() string {
	return a.String()
}

func (a addr) String() string {
	return "test"
}

type communicator struct{}

func (c communicator) Start(context.Context, *packer.RemoteCmd) error {
	return errors.New("communicator not supported")
}

func (c communicator) Upload(string, io.Reader, *os.FileInfo) error {
	return errors.New("communicator not supported")
}

func (c communicator) UploadDir(dst string, src string, exclude []string) error {
	return errors.New("communicator not supported")
}

func (c communicator) Download(string, io.Writer) error {
	return errors.New("communicator not supported")
}

func (c communicator) DownloadDir(src string, dst string, exclude []string) error {
	return errors.New("communicator not supported")
}
