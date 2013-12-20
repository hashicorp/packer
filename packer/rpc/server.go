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
	DefaultBuildEndpoint                = "Build"
	DefaultBuilderEndpoint              = "Builder"
	DefaultCacheEndpoint                = "Cache"
	DefaultCommandEndpoint              = "Command"
	DefaultCommunicatorEndpoint         = "Communicator"
	DefaultEnvironmentEndpoint          = "Environment"
	DefaultHookEndpoint                 = "Hook"
	DefaultPostProcessorEndpoint        = "PostProcessor"
	DefaultProvisionerEndpoint          = "Provisioner"
	DefaultUiEndpoint                   = "Ui"
)

// Server represents an RPC server for Packer. This must be paired on
// the other side with a Client.
type Server struct {
	mux      *MuxConn
	streamId uint32
	server   *rpc.Server
	closeMux bool
}

// NewServer returns a new Packer RPC server.
func NewServer(conn io.ReadWriteCloser) *Server {
	result := NewServerWithMux(NewMuxConn(conn, 1), 0)
	result.closeMux = true
	return result
}

func NewServerWithMux(mux *MuxConn, streamId uint32) *Server {
	return &Server{
		mux:      mux,
		streamId: streamId,
		server:   rpc.NewServer(),
		closeMux: false,
	}
}

func (s *Server) Close() error {
	if s.closeMux {
		log.Printf("[WARN] Shutting down mux conn in Server")
		return s.mux.Close()
	}

	return nil
}

func (s *Server) RegisterArtifact(a packer.Artifact) {
	s.server.RegisterName(DefaultArtifactEndpoint, &ArtifactServer{
		artifact: a,
	})
}

func (s *Server) RegisterBuild(b packer.Build) {
	s.server.RegisterName(DefaultBuildEndpoint, &BuildServer{
		build: b,
		mux:   s.mux,
	})
}

func (s *Server) RegisterBuilder(b packer.Builder) {
	s.server.RegisterName(DefaultBuilderEndpoint, &BuilderServer{
		builder: b,
		mux:     s.mux,
	})
}

func (s *Server) RegisterCache(c packer.Cache) {
	s.server.RegisterName(DefaultCacheEndpoint, &CacheServer{
		cache: c,
	})
}

func (s *Server) RegisterCommand(c packer.Command) {
	s.server.RegisterName(DefaultCommandEndpoint, &CommandServer{
		command: c,
		mux:     s.mux,
	})
}

func (s *Server) RegisterCommunicator(c packer.Communicator) {
	s.server.RegisterName(DefaultCommunicatorEndpoint, &CommunicatorServer{
		c:   c,
		mux: s.mux,
	})
}

func (s *Server) RegisterEnvironment(b packer.Environment) {
	s.server.RegisterName(DefaultEnvironmentEndpoint, &EnvironmentServer{
		env: b,
		mux: s.mux,
	})
}

func (s *Server) RegisterHook(h packer.Hook) {
	s.server.RegisterName(DefaultHookEndpoint, &HookServer{
		hook: h,
		mux:  s.mux,
	})
}

func (s *Server) RegisterPostProcessor(p packer.PostProcessor) {
	s.server.RegisterName(DefaultPostProcessorEndpoint, &PostProcessorServer{
		mux: s.mux,
		p:   p,
	})
}

func (s *Server) RegisterProvisioner(p packer.Provisioner) {
	s.server.RegisterName(DefaultProvisionerEndpoint, &ProvisionerServer{
		mux: s.mux,
		p:   p,
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
	stream, err := s.mux.Accept(s.streamId)
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
