package rpc

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net/rpc"
	"sync/atomic"
)

// Server represents an RPC server for Packer. This must be paired on
// the other side with a Client.
type Server struct {
	endpointId uint64
	rpcServer  *rpc.Server
}

// NewServer returns a new Packer RPC server.
func NewServer() *Server {
	return &Server{
		endpointId: 0,
		rpcServer:  rpc.NewServer(),
	}
}

func (s *Server) RegisterArtifact(a packer.Artifact) {
	s.registerComponent("Artifact", &ArtifactServer{a}, false)
}

// ServeConn serves a single connection over the RPC server. It is up
// to the caller to obtain a proper io.ReadWriteCloser.
func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	mux := NewMuxConn(conn)
	defer mux.Close()

	// Accept a connection on stream ID 0, which is always used for
	// normal client to server connections.
	stream, err := mux.Accept(0)
	if err != nil {
		log.Printf("[ERR] Error retrieving stream for serving: %s", err)
		return
	}

	s.rpcServer.ServeConn(stream)
}

// registerComponent registers a single Packer RPC component onto
// the RPC server. If id is true, then a unique ID number will be appended
// onto the end of the endpoint.
//
// The endpoint name is returned.
func (s *Server) registerComponent(name string, rcvr interface{}, id bool) string {
	endpoint := name
	if id {
		fmt.Sprintf("%s.%d", endpoint, atomic.AddUint64(&s.endpointId, 1))
	}

	s.rpcServer.RegisterName(endpoint, rcvr)
	return endpoint
}
