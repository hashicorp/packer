package rpc

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"net"
	"net/rpc"
)

// A Server is a Golang RPC server that has helper methods for automatically
// setting up the endpoints for Packer interfaces.
type Server struct {
	server *rpc.Server
	started bool
	doneChan chan bool
}

// Creates and returns a new Server.
func NewServer() *Server {
	return &Server{
		server: rpc.NewServer(),
		started: false,
	}
}

func (s *Server) Address() string {
	return ":2345"
}

func (s *Server) RegisterUi(ui packer.Ui) {
	s.server.RegisterName("Ui", &UiServer{ui})
}

func (s *Server) Start() error {
	if s.started {
		return errors.New("Server already started.")
	}

	// TODO: Address
	address := ":2345"

	// Mark that we started and setup the channel we'll use to mark exits
	s.started = true
	s.doneChan = make(chan bool)

	// Start the TCP listener and a goroutine responsible for cleaning up the
	// listener.
	listener, _ := net.Listen("tcp", address)
	go func() {
		<-s.doneChan
		listener.Close()
	}()

	// Start accepting connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}

			go s.server.ServeConn(conn)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.started {
		// TODO: There is a race condition here, we need to wait for
		// the listener to REALLY close.
		s.doneChan <- true
		s.started = false
		s.doneChan = nil
	}
}
