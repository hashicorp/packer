package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.BuilderFactory where the factory is actually
// executed over an RPC connection.
type BuilderFactory struct {
	client *rpc.Client
}

// BuilderFactoryServer wraps a packer.BuilderFactory and makes it exportable
// as part of a Golang RPC server.
type BuilderFactoryServer struct {
	bf packer.BuilderFactory
}

type BuilderFactoryCreateArgs struct {
	Name string
}

func (b *BuilderFactory) CreateBuilder(name string) packer.Builder {
	var reply string

	b.client.Call("BuilderFactory.CreateBuilder", &BuilderFactoryCreateArgs{name}, &reply)

	// TODO: error handling
	client, _ := rpc.Dial("tcp", reply)
	return &Builder{client}
}

func (b *BuilderFactoryServer) CreateBuilder(args *BuilderFactoryCreateArgs, reply *string) error {
	// Get the actual builder response
	builder := b.bf.CreateBuilder(args.Name)

	// Now we wrap that back up into a server, and send it on backwards.
	server := NewServer()
	server.RegisterBuilder(builder)
	server.StartSingle()

	// Set the reply to the address of the sever
	*reply = server.Address()
	return nil
}
