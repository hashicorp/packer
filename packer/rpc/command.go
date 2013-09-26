package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A Command is an implementation of the packer.Command interface where the
// command is actually executed over an RPC connection.
type command struct {
	client *rpc.Client
}

// A CommandServer wraps a packer.Command and makes it exportable as part
// of a Golang RPC server.
type CommandServer struct {
	command packer.Command
}

type CommandRunArgs struct {
	RPCAddress string
	Args       []string
}

type CommandSynopsisArgs byte

func Command(client *rpc.Client) *command {
	return &command{client}
}

func (c *command) Help() (result string) {
	err := c.client.Call("Command.Help", new(interface{}), &result)
	if err != nil {
		panic(err)
	}

	return
}

func (c *command) Run(env packer.Environment, args []string) (result int) {
	// Create and start the server for the Environment
	server := rpc.NewServer()
	RegisterEnvironment(server, env)

	rpcArgs := &CommandRunArgs{serveSingleConn(server), args}
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
	client, err := rpcDial(args.RPCAddress)
	if err != nil {
		return err
	}

	env := &Environment{client}

	*reply = c.command.Run(env, args.Args)
	return nil
}

func (c *CommandServer) Synopsis(args *CommandSynopsisArgs, reply *string) error {
	*reply = c.command.Synopsis()
	return nil
}
