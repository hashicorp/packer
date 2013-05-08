package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A ClientCommand is an implementation of the Command interface where the
// command is actually executed over an RPC connection.
type ClientCommand struct {
	client *rpc.Client
}

// A ServerCommand wraps a Command and makes it exportable as part
// of a Golang RPC server.
type ServerCommand struct {
	command packer.Command
}

type CommandRunArgs struct {
	RPCAddress string
	Args []string
}

type CommandSynopsisArgs byte

func Command(client *rpc.Client) *ClientCommand {
	return &ClientCommand{client}
}

func (c *ClientCommand) Run(env packer.Environment, args []string) (result int) {
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

func (c *ClientCommand) Synopsis() (result string) {
	err := c.client.Call("Command.Synopsis", CommandSynopsisArgs(0), &result)
	if err != nil {
		panic(err)
	}

	return
}

func (c *ServerCommand) Run(args *CommandRunArgs, reply *int) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	env := &Environment{client}

	*reply = c.command.Run(env, args.Args)
	return nil
}

func (c *ServerCommand) Synopsis(args *CommandSynopsisArgs, reply *string) error {
	*reply = c.command.Synopsis()
	return nil
}
