package rpc

import (
	"io"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/ugorji/go/codec"
)

// Client is the client end that communicates with a Packer RPC server.
// Establishing a connection is up to the user. The Client can communicate over
// any ReadWriteCloser. In Packer, each "plugin" (builder, provisioner,
// and post-processor) creates and launches a server. The the packer "core"
// creates and uses the client.
type Client struct {
	mux      *muxBroker
	client   *rpc.Client
	closeMux bool
}

func NewClient(rwc io.ReadWriteCloser) (*Client, error) {
	mux, err := newMuxBrokerClient(rwc)
	if err != nil {
		return nil, err
	}
	go mux.Run()

	result, err := newClientWithMux(mux, 0)
	if err != nil {
		mux.Close()
		return nil, err
	}

	result.closeMux = true
	return result, err
}

func newClientWithMux(mux *muxBroker, streamId uint32) (*Client, error) {
	clientConn, err := mux.Dial(streamId)
	if err != nil {
		return nil, err
	}

	h := &codec.MsgpackHandle{
		WriteExt: true,
	}
	clientCodec := codec.GoRpc.ClientCodec(clientConn, h)

	return &Client{
		mux:      mux,
		client:   rpc.NewClientWithCodec(clientCodec),
		closeMux: false,
	}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return err
	}

	if c.closeMux {
		log.Printf("[WARN] Client is closing mux")
		return c.mux.Close()
	}

	return nil
}

func (c *Client) Artifact() packer.Artifact {
	return &artifact{
		commonClient: commonClient{
			endpoint: DefaultArtifactEndpoint,
			client:   c.client,
		},
	}
}

func (c *Client) Build() packer.Build {
	return &build{
		commonClient: commonClient{
			endpoint: DefaultBuildEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Builder() packer.Builder {
	return &builder{
		commonClient: commonClient{
			endpoint: DefaultBuilderEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Communicator() packer.Communicator {
	return &communicator{
		commonClient: commonClient{
			endpoint: DefaultCommunicatorEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Hook() packer.Hook {
	return &hook{
		commonClient: commonClient{
			endpoint: DefaultHookEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) PostProcessor() packer.PostProcessor {
	return &postProcessor{
		commonClient: commonClient{
			endpoint: DefaultPostProcessorEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Provisioner() packer.Provisioner {
	return &provisioner{
		commonClient: commonClient{
			endpoint: DefaultProvisionerEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Datasource() packer.Datasource {
	return &datasource{
		commonClient: commonClient{
			endpoint: DefaultDatasourceEndpoint,
			client:   c.client,
			mux:      c.mux,
		},
	}
}

func (c *Client) Ui() packer.Ui {
	return &Ui{
		commonClient: commonClient{
			endpoint: DefaultUiEndpoint,
			client:   c.client,
		},
		endpoint: DefaultUiEndpoint,
	}
}
