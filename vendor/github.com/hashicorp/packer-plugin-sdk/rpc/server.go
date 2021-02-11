package rpc

import (
	"io"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer-plugin-sdk/packer"
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
	DefaultDatasourceEndpoint           = "Datasource"
	DefaultUiEndpoint                   = "Ui"
)

// PluginServer represents an RPC server for Packer. This must be paired on the
// other side with a PluginClient. In Packer, each "plugin" (builder,
// provisioner, and post-processor) creates and launches a server. The client
// created and used by the packer "core"
type PluginServer struct {
	mux      *muxBroker
	streamId uint32
	server   *rpc.Server
	closeMux bool
}

// NewServer returns a new Packer RPC server.
func NewServer(conn io.ReadWriteCloser) (*PluginServer, error) {
	mux, err := newMuxBrokerServer(conn)
	if err != nil {
		return nil, err
	}
	result := newServerWithMux(mux, 0)
	result.closeMux = true
	go mux.Run()
	return result, nil
}

func newServerWithMux(mux *muxBroker, streamId uint32) *PluginServer {
	return &PluginServer{
		mux:      mux,
		streamId: streamId,
		server:   rpc.NewServer(),
		closeMux: false,
	}
}

func (s *PluginServer) Close() error {
	if s.closeMux {
		log.Printf("[WARN] Shutting down mux conn in Server")
		return s.mux.Close()
	}

	return nil
}

func (s *PluginServer) RegisterArtifact(a packer.Artifact) error {
	return s.server.RegisterName(DefaultArtifactEndpoint, &ArtifactServer{
		artifact: a,
	})
}

func (s *PluginServer) RegisterBuild(b packer.Build) error {
	return s.server.RegisterName(DefaultBuildEndpoint, &BuildServer{
		build: b,
		mux:   s.mux,
	})
}

func (s *PluginServer) RegisterBuilder(b packer.Builder) error {
	return s.server.RegisterName(DefaultBuilderEndpoint, &BuilderServer{
		commonServer: commonServer{
			selfConfigurable: b,
			mux:              s.mux,
		},
		builder: b,
	})
}

func (s *PluginServer) RegisterCommunicator(c packer.Communicator) error {
	return s.server.RegisterName(DefaultCommunicatorEndpoint, &CommunicatorServer{
		c: c,
		commonServer: commonServer{
			mux: s.mux,
		},
	})
}

func (s *PluginServer) RegisterHook(h packer.Hook) error {
	return s.server.RegisterName(DefaultHookEndpoint, &HookServer{
		hook: h,
		mux:  s.mux,
	})
}

func (s *PluginServer) RegisterPostProcessor(p packer.PostProcessor) error {
	return s.server.RegisterName(DefaultPostProcessorEndpoint, &PostProcessorServer{
		commonServer: commonServer{
			selfConfigurable: p,
			mux:              s.mux,
		},
		p: p,
	})
}

func (s *PluginServer) RegisterProvisioner(p packer.Provisioner) error {
	return s.server.RegisterName(DefaultProvisionerEndpoint, &ProvisionerServer{
		commonServer: commonServer{
			selfConfigurable: p,
			mux:              s.mux,
		},
		p: p,
	})
}

func (s *PluginServer) RegisterDatasource(d packer.Datasource) error {
	return s.server.RegisterName(DefaultDatasourceEndpoint, &DatasourceServer{
		commonServer: commonServer{
			selfConfigurable: d,
			mux:              s.mux,
		},
		d: d,
	})
}

func (s *PluginServer) RegisterUi(ui packer.Ui) error {
	return s.server.RegisterName(DefaultUiEndpoint, &UiServer{
		ui:       ui,
		register: s.server.RegisterName,
	})
}

// ServeConn serves a single connection over the RPC server. It is up
// to the caller to obtain a proper io.ReadWriteCloser.
func (s *PluginServer) Serve() {
	// Accept a connection on stream ID 0, which is always used for
	// normal client to server connections.
	stream, err := s.mux.Accept(s.streamId)
	if err != nil {
		log.Printf("[ERR] Error retrieving stream for serving: %s", err)
		return
	}
	defer stream.Close()

	h := &codec.MsgpackHandle{
		WriteExt: true,
	}
	rpcCodec := codec.GoRpc.ServerCodec(stream, h)
	s.server.ServeCodec(rpcCodec)
}
