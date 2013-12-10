package rpc

import (
	"github.com/mitchellh/packer/packer"
	"io"
	"net/rpc"
)

// Client is the client end that communicates with a Packer RPC server.
// Establishing a connection is up to the user, the Client can just
// communicate over any ReadWriteCloser.
type Client struct {
	mux    *MuxConn
	client *rpc.Client
}

func NewClient(rwc io.ReadWriteCloser) (*Client, error) {
	return NewClientWithMux(NewMuxConn(rwc), 0)
}

func NewClientWithMux(mux *MuxConn, streamId uint32) (*Client, error) {
	clientConn, err := mux.Dial(streamId)
	if err != nil {
		return nil, err
	}

	return &Client{
		mux:    mux,
		client: rpc.NewClient(clientConn),
	}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Artifact() packer.Artifact {
	return &artifact{
		client:   c.client,
		endpoint: DefaultArtifactEndpoint,
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
