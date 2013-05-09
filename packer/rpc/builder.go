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

func (b *builder) Prepare(config interface{}) {
	b.client.Call("Builder.Prepare", &BuilderPrepareArgs{config}, new(interface{}))
}

func (b *builder) Run(build packer.Build, ui packer.Ui) {
	// Create and start the server for the Build and UI
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterBuild(server, build)
	RegisterUi(server, ui)

	args := &BuilderRunArgs{serveSingleConn(server)}
	b.client.Call("Builder.Run", args, new(interface{}))
}

func (b *BuilderServer) Prepare(args *BuilderPrepareArgs, reply *interface{}) error {
	b.builder.Prepare(args.Config)
	*reply = nil
	return nil
}

func (b *BuilderServer) Run(args *BuilderRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	build := &Build{client}
	ui := &Ui{client}
	b.builder.Run(build, ui)

	*reply = nil
	return nil
}
