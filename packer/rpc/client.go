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
	// Create the MuxConn around the RWC and get the client to server stream.
	// This is the primary stream that we use to communicate with the
	// remote RPC server. On the remote side Server.ServeConn also listens
	// on this stream ID.
	mux := NewMuxConn(rwc)
	stream, err := mux.Dial(0)
	if err != nil {
		return nil, err
	}

	return &Client{
		mux:    mux,
		client: rpc.NewClient(stream),
	}, nil
}

func (c *Client) Artifact() packer.Artifact {
	return &artifact{
		client: c.client,
	}
}
