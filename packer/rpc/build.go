package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Build where the build is actually executed
// over an RPC connection.
type Build struct {
	client *rpc.Client
}

// BuildServer wraps a packer.Build implementation and makes it exportable
// as part of a Golang RPC server.
type BuildServer struct {
	build packer.Build
}

type BuildPrepareArgs interface{}

type BuildRunArgs struct {
	UiRPCAddress string
}

func (b *Build) Prepare() {
	b.client.Call("Build.Prepare", new(interface{}), new(interface{}))
}

func (b *Build) Run(ui packer.Ui) {
	// Create and start the server for the UI
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterUi(server, ui)
	args := &BuildRunArgs{serveSingleConn(server)}
	b.client.Call("Build.Run", args, new(interface{}))
}

func (b *BuildServer) Prepare(args *BuildPrepareArgs, reply *interface{}) error {
	b.build.Prepare()

	*reply = nil
	return nil
}

func (b *BuildServer) Run(args *BuildRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.UiRPCAddress)
	if err != nil {
		return err
	}

	b.build.Run(&Ui{client})

	*reply = nil
	return nil
}
