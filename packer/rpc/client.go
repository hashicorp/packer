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
	server *rpc.Server
}

func NewClient(rwc io.ReadWriteCloser) (*Client, error) {
	// Create the MuxConn around the RWC and get the client to server stream.
	// This is the primary stream that we use to communicate with the
	// remote RPC server. On the remote side Server.ServeConn also listens
	// on this stream ID.
	mux := NewMuxConn(rwc)
	clientConn, err := mux.Dial(0)
	if err != nil {
		mux.Close()
		return nil, err
	}

	// Accept connection ID 1 which is what the remote end uses to
	// be an RPC client back to us so we can even serve some objects.
	serverConn, err := mux.Accept(1)
	if err != nil {
		mux.Close()
		return nil, err
	}

	// Start our RPC server on this end
	server := rpc.NewServer()
	go server.ServeConn(serverConn)

	return &Client{
		mux:    mux,
		client: rpc.NewClient(clientConn),
		server: server,
	}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return err
	}

	return c.mux.Close()
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

func (c *Client) PostProcessor() packer.PostProcessor {
	return &postProcessor{
		client: c.client,
		server: c.server,
	}
}
