package rpc

import (
	"github.com/mitchellh/packer/packer"
	"log"
	"net/rpc"
)

// A Environment is an implementation of the packer.Environment interface
// where the actual environment is executed over an RPC connection.
type Environment struct {
	client *rpc.Client
	mux    *MuxConn
}

// A EnvironmentServer wraps a packer.Environment and makes it exportable
// as part of a Golang RPC server.
type EnvironmentServer struct {
	env packer.Environment
	mux *MuxConn
}

type EnvironmentCliArgs struct {
	Args []string
}

func (e *Environment) Builder(name string) (b packer.Builder, err error) {
	var streamId uint32
	err = e.client.Call("Environment.Builder", name, &streamId)
	if err != nil {
		return
	}

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		return nil, err
	}
	b = client.Builder()
	return
}

func (e *Environment) Cache() packer.Cache {
	var streamId uint32
	if err := e.client.Call("Environment.Cache", new(interface{}), &streamId); err != nil {
		panic(err)
	}

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		log.Printf("[ERR] Error getting cache client: %s", err)
		return nil
	}
	return client.Cache()
}

func (e *Environment) Cli(args []string) (result int, err error) {
	rpcArgs := &EnvironmentCliArgs{args}
	err = e.client.Call("Environment.Cli", rpcArgs, &result)
	return
}

func (e *Environment) Hook(name string) (h packer.Hook, err error) {
	var streamId uint32
	err = e.client.Call("Environment.Hook", name, &streamId)
	if err != nil {
		return
	}

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		return nil, err
	}
	return client.Hook(), nil
}

func (e *Environment) PostProcessor(name string) (p packer.PostProcessor, err error) {
	var streamId uint32
	err = e.client.Call("Environment.PostProcessor", name, &streamId)
	if err != nil {
		return
	}

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		return nil, err
	}
	p = client.PostProcessor()
	return
}

func (e *Environment) Provisioner(name string) (p packer.Provisioner, err error) {
	var streamId uint32
	err = e.client.Call("Environment.Provisioner", name, &streamId)
	if err != nil {
		return
	}

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		return nil, err
	}
	p = client.Provisioner()
	return
}

func (e *Environment) Ui() packer.Ui {
	var streamId uint32
	e.client.Call("Environment.Ui", new(interface{}), &streamId)

	client, err := newClientWithMux(e.mux, streamId)
	if err != nil {
		log.Printf("[ERR] Error connecting to Ui: %s", err)
		return nil
	}
	return client.Ui()
}

func (e *EnvironmentServer) Builder(name string, reply *uint32) error {
	builder, err := e.env.Builder(name)
	if err != nil {
		return NewBasicError(err)
	}

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterBuilder(builder)
	go server.Serve()
	return nil
}

func (e *EnvironmentServer) Cache(args *interface{}, reply *uint32) error {
	cache := e.env.Cache()

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterCache(cache)
	go server.Serve()
	return nil
}

func (e *EnvironmentServer) Cli(args *EnvironmentCliArgs, reply *int) (err error) {
	*reply, err = e.env.Cli(args.Args)
	return
}

func (e *EnvironmentServer) Hook(name string, reply *uint32) error {
	hook, err := e.env.Hook(name)
	if err != nil {
		return NewBasicError(err)
	}

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterHook(hook)
	go server.Serve()
	return nil
}

func (e *EnvironmentServer) PostProcessor(name string, reply *uint32) error {
	pp, err := e.env.PostProcessor(name)
	if err != nil {
		return NewBasicError(err)
	}

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterPostProcessor(pp)
	go server.Serve()
	return nil
}

func (e *EnvironmentServer) Provisioner(name string, reply *uint32) error {
	prov, err := e.env.Provisioner(name)
	if err != nil {
		return NewBasicError(err)
	}

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterProvisioner(prov)
	go server.Serve()
	return nil
}

func (e *EnvironmentServer) Ui(args *interface{}, reply *uint32) error {
	ui := e.env.Ui()

	*reply = e.mux.NextId()
	server := newServerWithMux(e.mux, *reply)
	server.RegisterUi(ui)
	go server.Serve()
	return nil
}
