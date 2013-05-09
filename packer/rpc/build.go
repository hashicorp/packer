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

func (b *Build) Name() (result string) {
	b.client.Call("Build.Name", new(interface{}), &result)
	return
}

func (b *Build) Prepare() (err error) {
	b.client.Call("Build.Prepare", new(interface{}), &err)
	return
}

func (b *Build) Run(ui packer.Ui) {
	// Create and start the server for the UI
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterUi(server, ui)
	args := &BuildRunArgs{serveSingleConn(server)}
	b.client.Call("Build.Run", args, new(interface{}))
}

func (b *BuildServer) Name(args *interface{}, reply*string) error {
	*reply = b.build.Name()
	return nil
}

func (b *BuildServer) Prepare(args *BuildPrepareArgs, reply *error) error {
	*reply = b.build.Prepare()
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
