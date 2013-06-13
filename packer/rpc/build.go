package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Build where the build is actually executed
// over an RPC connection.
type build struct {
	client *rpc.Client
}

// BuildServer wraps a packer.Build implementation and makes it exportable
// as part of a Golang RPC server.
type BuildServer struct {
	build packer.Build
}

type BuildRunArgs struct {
	UiRPCAddress string
}

func Build(client *rpc.Client) *build {
	return &build{client}
}

func (b *build) Name() (result string) {
	b.client.Call("Build.Name", new(interface{}), &result)
	return
}

func (b *build) Prepare() (err error) {
	if cerr := b.client.Call("Build.Prepare", new(interface{}), &err); cerr != nil {
		return cerr
	}

	return
}

func (b *build) Run(ui packer.Ui, cache packer.Cache) (packer.Artifact, error) {
	// Create and start the server for the UI
	server := rpc.NewServer()
	RegisterCache(server, cache)
	RegisterUi(server, ui)
	args := &BuildRunArgs{serveSingleConn(server)}

	var reply string
	if err := b.client.Call("Build.Run", args, &reply); err != nil {
		return nil, err
	}

	client, err := rpc.Dial("tcp", reply)
	if err != nil {
		return nil, err
	}

	return Artifact(client), nil
}

func (b *build) Cancel() {
	if err := b.client.Call("Build.Cancel", new(interface{}), new(interface{})); err != nil {
		panic(err)
	}
}

func (b *BuildServer) Name(args *interface{}, reply *string) error {
	*reply = b.build.Name()
	return nil
}

func (b *BuildServer) Prepare(args interface{}, reply *error) error {
	*reply = b.build.Prepare()
	return nil
}

func (b *BuildServer) Run(args *BuildRunArgs, reply *string) error {
	client, err := rpc.Dial("tcp", args.UiRPCAddress)
	if err != nil {
		return err
	}

	artifact, err := b.build.Run(&Ui{client}, Cache(client))
	if err != nil {
		return NewBasicError(err)
	}

	// Wrap the artifact
	server := rpc.NewServer()
	RegisterArtifact(server, artifact)

	*reply = serveSingleConn(server)
	return nil
}

func (b *BuildServer) Cancel(args *interface{}, reply *interface{}) error {
	b.build.Cancel()
	return nil
}
