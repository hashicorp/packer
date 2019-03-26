package rpc

import (
	baserpc "net/rpc"

	"github.com/keegancsmith/rpc"
)

type Server = rpc.Server
type Client = rpc.Client
type Request = rpc.Request
type Response = rpc.Response

type ClientCodec interface {
	WriteRequest(*baserpc.Request, interface{}) error
	ReadResponseHeader(*baserpc.Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

type ServerCodec interface {
	ReadRequestHeader(*baserpc.Request) error
	ReadRequestBody(interface{}) error
	WriteResponse(*baserpc.Response, interface{}) error

	// Close can be called multiple times and must be idempotent.
	Close() error
}

func NewClientWithCodec(codec rpc.ClientCodec) *Client {
	return rpc.NewClientWithCodec(codec)
}

func NewServer() *Server {
	return rpc.NewServer()
}
