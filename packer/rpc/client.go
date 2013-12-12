package rpc

import (
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net/rpc"
)

// Client is the client end that communicates with a Packer RPC server.
// Establishing a connection is up to the user, the Client can just
// communicate over any ReadWriteCloser.
type Client struct {
	mux      *MuxConn
	client   *rpc.Client
	closeMux bool
}

func NewClient(rwc io.ReadWriteCloser) (*Client, error) {
	result, err := NewClientWithMux(NewMuxConn(rwc), 0)
	if err != nil {
		return nil, err
	}

	result.closeMux = true
	return result, err
}

func NewClientWithMux(mux *MuxConn, streamId uint32) (*Client, error) {
	clientConn, err := mux.Dial(streamId)
	if err != nil {
		return nil, err
	}

	return &Client{
		mux:      mux,
		client:   rpc.NewClient(clientConn),
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
		client:   c.client,
		endpoint: DefaultArtifactEndpoint,
	}
}

func (c *Client) Build() packer.Build {
	return &build{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Builder() packer.Builder {
	return &builder{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Cache() packer.Cache {
	return &cache{
		client: c.client,
	}
}

func (c *Client) Command() packer.Command {
	return &command{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Communicator() packer.Communicator {
	return &communicator{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Environment() packer.Environment {
	return &Environment{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Hook() packer.Hook {
	return &hook{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) PostProcessor() packer.PostProcessor {
	return &postProcessor{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Provisioner() packer.Provisioner {
	return &provisioner{
		client: c.client,
		mux:    c.mux,
	}
}

func (c *Client) Ui() packer.Ui {
	return &Ui{
		client:   c.client,
		endpoint: DefaultUiEndpoint,
	}
}
