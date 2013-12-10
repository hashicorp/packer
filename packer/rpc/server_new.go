package rpc

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net/rpc"
	"sync/atomic"
)

var endpointId uint64

const (
	DefaultArtifactEndpoint      string = "Artifact"
	DefaultCacheEndpoint                = "Cache"
	DefaultPostProcessorEndpoint        = "PostProcessor"
	DefaultUiEndpoint                   = "Ui"
)

// Server represents an RPC server for Packer. This must be paired on
// the other side with a Client.
type Server struct {
	mux    *MuxConn
	server *rpc.Server
}

// NewServer returns a new Packer RPC server.
func NewServer(conn io.ReadWriteCloser) *Server {
	return &Server{
		mux:    NewMuxConn(conn),
		server: rpc.NewServer(),
	}
}

func (s *Server) Close() error {
	return s.mux.Close()
}

func (s *Server) RegisterArtifact(a packer.Artifact) {
	s.server.RegisterName(DefaultArtifactEndpoint, &ArtifactServer{
		artifact: a,
	})
}

func (s *Server) RegisterCache(c packer.Cache) {
	s.server.RegisterName(DefaultCacheEndpoint, &CacheServer{
		cache: c,
	})
}

func (s *Server) RegisterPostProcessor(p packer.PostProcessor) {
	s.server.RegisterName(DefaultPostProcessorEndpoint, &PostProcessorServer{
		p: p,
	})
}

func (s *Server) RegisterUi(ui packer.Ui) {
	s.server.RegisterName(DefaultUiEndpoint, &UiServer{
		ui: ui,
	})
}

// ServeConn serves a single connection over the RPC server. It is up
// to the caller to obtain a proper io.ReadWriteCloser.
func (s *Server) Serve() {
	// Accept a connection on stream ID 0, which is always used for
	// normal client to server connections.
	stream, err := s.mux.Accept(0)
	defer stream.Close()
	if err != nil {
		log.Printf("[ERR] Error retrieving stream for serving: %s", err)
		return
	}

	s.server.ServeConn(stream)
}

// registerComponent registers a single Packer RPC component onto
// the RPC server. If id is true, then a unique ID number will be appended
// onto the end of the endpoint.
//
// The endpoint name is returned.
func registerComponent(server *rpc.Server, name string, rcvr interface{}, id bool) string {
	endpoint := name
	if id {
		fmt.Sprintf("%s.%d", endpoint, atomic.AddUint64(&endpointId, 1))
	}

	server.RegisterName(endpoint, rcvr)
	return endpoint
}
