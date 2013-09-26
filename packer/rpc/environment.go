package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A Environment is an implementation of the packer.Environment interface
// where the actual environment is executed over an RPC connection.
type Environment struct {
	client *rpc.Client
}

// A EnvironmentServer wraps a packer.Environment and makes it exportable
// as part of a Golang RPC server.
type EnvironmentServer struct {
	env packer.Environment
}

type EnvironmentCliArgs struct {
	Args []string
}

func (e *Environment) Builder(name string) (b packer.Builder, err error) {
	var reply string
	err = e.client.Call("Environment.Builder", name, &reply)
	if err != nil {
		return
	}

	client, err := rpcDial(reply)
	if err != nil {
		return
	}

	b = Builder(client)
	return
}

func (e *Environment) Cache() packer.Cache {
	var reply string
	if err := e.client.Call("Environment.Cache", new(interface{}), &reply); err != nil {
		panic(err)
	}

	client, err := rpcDial(reply)
	if err != nil {
		panic(err)
	}

	return Cache(client)
}

func (e *Environment) Cli(args []string) (result int, err error) {
	rpcArgs := &EnvironmentCliArgs{args}
	err = e.client.Call("Environment.Cli", rpcArgs, &result)
	return
}

func (e *Environment) Hook(name string) (h packer.Hook, err error) {
	var reply string
	err = e.client.Call("Environment.Hook", name, &reply)
	if err != nil {
		return
	}

	client, err := rpcDial(reply)
	if err != nil {
		return
	}

	h = Hook(client)
	return
}

func (e *Environment) PostProcessor(name string) (p packer.PostProcessor, err error) {
	var reply string
	err = e.client.Call("Environment.PostProcessor", name, &reply)
	if err != nil {
		return
	}

	client, err := rpcDial(reply)
	if err != nil {
		return
	}

	p = PostProcessor(client)
	return
}

func (e *Environment) Provisioner(name string) (p packer.Provisioner, err error) {
	var reply string
	err = e.client.Call("Environment.Provisioner", name, &reply)
	if err != nil {
		return
	}

	client, err := rpcDial(reply)
	if err != nil {
		return
	}

	p = Provisioner(client)
	return
}

func (e *Environment) Ui() packer.Ui {
	var reply string
	e.client.Call("Environment.Ui", new(interface{}), &reply)

	client, err := rpcDial(reply)
	if err != nil {
		panic(err)
	}

	return &Ui{client}
}

func (e *EnvironmentServer) Builder(name *string, reply *string) error {
	builder, err := e.env.Builder(*name)
	if err != nil {
		return err
	}

	// Wrap it
	server := rpc.NewServer()
	RegisterBuilder(server, builder)

	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) Cache(args *interface{}, reply *string) error {
	cache := e.env.Cache()

	server := rpc.NewServer()
	RegisterCache(server, cache)
	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) Cli(args *EnvironmentCliArgs, reply *int) (err error) {
	*reply, err = e.env.Cli(args.Args)
	return
}

func (e *EnvironmentServer) Hook(name *string, reply *string) error {
	hook, err := e.env.Hook(*name)
	if err != nil {
		return err
	}

	// Wrap it
	server := rpc.NewServer()
	RegisterHook(server, hook)

	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) PostProcessor(name *string, reply *string) error {
	pp, err := e.env.PostProcessor(*name)
	if err != nil {
		return err
	}

	server := rpc.NewServer()
	RegisterPostProcessor(server, pp)

	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) Provisioner(name *string, reply *string) error {
	prov, err := e.env.Provisioner(*name)
	if err != nil {
		return err
	}

	server := rpc.NewServer()
	RegisterProvisioner(server, prov)

	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) Ui(args *interface{}, reply *string) error {
	ui := e.env.Ui()

	// Wrap it
	server := rpc.NewServer()
	RegisterUi(server, ui)

	*reply = serveSingleConn(server)
	return nil
}
