package rpc

import (
	"io"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer/packer"
	"github.com/ugorji/go/codec"
)

const (
	DefaultArtifactEndpoint      string = "Artifact"
	DefaultBuildEndpoint                = "Build"
	DefaultBuilderEndpoint              = "Builder"
	DefaultCacheEndpoint                = "Cache"
	DefaultCommandEndpoint              = "Command"
	DefaultCommunicatorEndpoint         = "Communicator"
	DefaultHookEndpoint                 = "Hook"
	DefaultPostProcessorEndpoint        = "PostProcessor"
	DefaultProvisionerEndpoint          = "Provisioner"
	DefaultUiEndpoint                   = "Ui"
)

// Server represents an RPC server for Packer. This must be paired on
// the other side with a Client.
type Server struct {
	mux      *muxBroker
	streamId uint32
	server   *rpc.Server
	closeMux bool
}

// NewServer returns a new Packer RPC server.
func NewServer(conn io.ReadWriteCloser) (*Server, error) {
	mux, err := newMuxBrokerServer(conn)
	if err != nil {
		return nil, err
	}
	result := newServerWithMux(mux, 0)
	result.closeMux = true
	go mux.Run()
	return result, nil
}

func newServerWithMux(mux *muxBroker, streamId uint32) *Server {
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

func (s *Server) RegisterArtifact(a packer.Artifact) error {
	return s.server.RegisterName(DefaultArtifactEndpoint, &ArtifactServer{
		artifact: a,
	})
}

func (s *Server) RegisterBuild(b packer.Build) error {
	return s.server.RegisterName(DefaultBuildEndpoint, &BuildServer{
		build: b,
		mux:   s.mux,
	})
}

func (s *Server) RegisterBuilder(b packer.Builder) error {
	return s.server.RegisterName(DefaultBuilderEndpoint, &BuilderServer{
		commonServer: commonServer{
			selfConfigurable: b,
			mux:              s.mux,
		},
		builder: b,
	})
}

func (s *Server) RegisterCommunicator(c packer.Communicator) error {
	return s.server.RegisterName(DefaultCommunicatorEndpoint, &CommunicatorServer{
		c: c,
		commonServer: commonServer{
			mux: s.mux,
		},
	})
}

func (s *Server) RegisterHook(h packer.Hook) error {
	return s.server.RegisterName(DefaultHookEndpoint, &HookServer{
		hook: h,
		mux:  s.mux,
	})
}

func (s *Server) RegisterPostProcessor(p packer.PostProcessor) error {
	return s.server.RegisterName(DefaultPostProcessorEndpoint, &PostProcessorServer{
		commonServer: commonServer{
			selfConfigurable: p,
			mux:              s.mux,
		},
		p: p,
	})
}

func (s *Server) RegisterProvisioner(p packer.Provisioner) error {
	return s.server.RegisterName(DefaultProvisionerEndpoint, &ProvisionerServer{
		commonServer: commonServer{
			selfConfigurable: p,
			mux:              s.mux,
		},
		p: p,
	})
}

func (s *Server) RegisterUi(ui packer.Ui) error {
	return s.server.RegisterName(DefaultUiEndpoint, &UiServer{
		ui:       ui,
		register: s.server.RegisterName,
	})
}

// ServeConn serves a single connection over the RPC server. It is up
// to the caller to obtain a proper io.ReadWriteCloser.
func (s *Server) Serve() {
	// Accept a connection on stream ID 0, which is always used for
	// normal client to server connections.
	stream, err := s.mux.Accept(s.streamId)
	if err != nil {
		log.Printf("[ERR] Error retrieving stream for serving: %s", err)
		return
	}
	defer stream.Close()

	h := &codec.MsgpackHandle{
		RawToString: true,
		WriteExt:    true,
	}
	rpcCodec := codec.GoRpc.ServerCodec(stream, h)
	s.server.ServeCodec(rpcCodec)
}
