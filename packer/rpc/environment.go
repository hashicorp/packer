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

func (e *Environment) Cli(args []string) (result int) {
	rpcArgs := &EnvironmentCliArgs{args}
	e.client.Call("Environment.Cli", rpcArgs, &result)
	return
}

func (e *Environment) Ui() packer.Ui {
	var reply string
	e.client.Call("Environment.Ui", new(interface{}), &reply)

	// TODO: error handling
	client, _ := rpc.Dial("tcp", reply)
	return &Ui{client}
}

func (e *EnvironmentServer) Cli(args *EnvironmentCliArgs, reply *int) error {
	*reply = e.env.Cli(args.Args)
	return nil
}

func (e *EnvironmentServer) Ui(args *interface{}, reply *string) error {
	ui := e.env.Ui()

	// Wrap it
	server := NewServer()
	server.RegisterUi(ui)
	server.StartSingle()

	*reply = server.Address()
	return nil
}
