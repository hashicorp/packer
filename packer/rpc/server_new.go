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
	DefaultArtifactEndpoint string = "Artifact"
)

// Server represents an RPC server for Packer. This must be paired on
// the other side with a Client.
type Server struct {
	components map[string]interface{}
}

// NewServer returns a new Packer RPC server.
func NewServer() *Server {
	return &Server{
		components: make(map[string]interface{}),
	}
}

func (s *Server) RegisterArtifact(a packer.Artifact) {
	s.components[DefaultArtifactEndpoint] = a
}

func (s *Server) RegisterCache(c packer.Cache) {
	s.components["Cache"] = c
}

func (s *Server) RegisterPostProcessor(p packer.PostProcessor) {
	s.components["PostProcessor"] = p
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

	clientConn, err := mux.Dial(1)
	if err != nil {
		log.Printf("[ERR] Error connecting to client stream: %s", err)
		return
	}
	client := rpc.NewClient(clientConn)

	// Create the RPC server
	server := rpc.NewServer()
	for endpoint, iface := range s.components {
		var endpointVal interface{}

		switch v := iface.(type) {
		case packer.Artifact:
			endpointVal = &ArtifactServer{
				artifact: v,
			}
		case packer.Cache:
			endpointVal = &CacheServer{
				cache: v,
			}
		case packer.PostProcessor:
			endpointVal = &PostProcessorServer{
				client: client,
				server: server,
				p: v,
			}
		default:
			log.Printf("[ERR] Unknown component for endpoint: %s", endpoint)
			return
		}

		registerComponent(server, endpoint, endpointVal, false)
	}

	server.ServeConn(stream)
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
