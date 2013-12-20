package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A Command is an implementation of the packer.Command interface where the
// command is actually executed over an RPC connection.
type command struct {
	client *rpc.Client
	mux    *MuxConn
}

// A CommandServer wraps a packer.Command and makes it exportable as part
// of a Golang RPC server.
type CommandServer struct {
	command packer.Command
	mux     *MuxConn
}

type CommandRunArgs struct {
	Args     []string
	StreamId uint32
}

type CommandSynopsisArgs byte

func (c *command) Help() (result string) {
	err := c.client.Call("Command.Help", new(interface{}), &result)
	if err != nil {
		panic(err)
	}

	return
}

func (c *command) Run(env packer.Environment, args []string) (result int) {
	nextId := c.mux.NextId()
	server := newServerWithMux(c.mux, nextId)
	server.RegisterEnvironment(env)
	go server.Serve()

	rpcArgs := &CommandRunArgs{
		Args:     args,
		StreamId: nextId,
	}
	err := c.client.Call("Command.Run", rpcArgs, &result)
	if err != nil {
		panic(err)
	}

	return
}

func (c *command) Synopsis() (result string) {
	err := c.client.Call("Command.Synopsis", CommandSynopsisArgs(0), &result)
	if err != nil {
		panic(err)
	}

	return
}

func (c *CommandServer) Help(args *interface{}, reply *string) error {
	*reply = c.command.Help()
	return nil
}

func (c *CommandServer) Run(args *CommandRunArgs, reply *int) error {
	client, err := newClientWithMux(c.mux, args.StreamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	*reply = c.command.Run(client.Environment(), args.Args)
	return nil
}

func (c *CommandServer) Synopsis(args *CommandSynopsisArgs, reply *string) error {
	*reply = c.command.Synopsis()
	return nil
}
