package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Builder where the builder is actually executed
// over an RPC connection.
type builder struct {
	client *rpc.Client
}

// BuilderServer wraps a packer.Builder implementation and makes it exportable
// as part of a Golang RPC server.
type BuilderServer struct {
	builder packer.Builder
}

type BuilderPrepareArgs struct {
	Config interface{}
}

type BuilderRunArgs struct {
	RPCAddress string
}

func Builder(client *rpc.Client) *builder {
	return &builder{client}
}

func (b *builder) Prepare(config interface{}) (err error) {
	cerr := b.client.Call("Builder.Prepare", &BuilderPrepareArgs{config}, &err)
	if cerr != nil {
		err = cerr
	}

	return
}

func (b *builder) Run(ui packer.Ui, hook packer.Hook) {
	// Create and start the server for the Build and UI
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterUi(server, ui)
	RegisterHook(server, hook)

	args := &BuilderRunArgs{serveSingleConn(server)}
	b.client.Call("Builder.Run", args, new(interface{}))
}

func (b *BuilderServer) Prepare(args *BuilderPrepareArgs, reply *error) error {
	err := b.builder.Prepare(args.Config)
	if err != nil {
		*reply = NewBasicError(err)
	}

	return nil
}

func (b *BuilderServer) Run(args *BuilderRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	hook := Hook(client)
	ui := &Ui{client}
	b.builder.Run(ui, hook)

	*reply = nil
	return nil
}
