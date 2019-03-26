package codec

import (
	"io"
	baserpc "net/rpc"

	"github.com/hashicorp/packer/common/net/rpc"
	"github.com/ugorji/go/codec"
)

type Handle = codec.Handle

var msgpackHandle = &codec.MsgpackHandle{
	RawToString: true,
	WriteExt:    true,
}

type serverCodec struct {
	baserpc.ServerCodec
}

func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	return s.ServerCodec.ReadRequestHeader(r)
}

func (s *serverCodec) WriteResponse(r *baserpc.Response, v interface{}) error {
	return s.ServerCodec.WriteResponse(r, v)
}

func MsgpackServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	c := codec.GoRpc.ServerCodec(conn, msgpackHandle)
	return &serverCodec{c}
}

type clientCodec struct {
	baserpc.ClientCodec
}

func (c *clientCodec) WriteRequest(req *baserpc.Request, v interface{}) error {
	return c.ClientCodec.WriteRequest(req, v)
}

func (c *clientCodec) ReadResponseHeader(res *rpc.Response) error {
	r := baserpc.Response(*res)
	return c.ClientCodec.ReadResponseHeader()
}

func MsgpackClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	c := codec.GoRpc.ClientCodec(conn, msgpackHandle)
	return &clientCodec{c}
}
