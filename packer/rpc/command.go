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
	Env packer.Environment
	Args []string
}

type CommandSynopsisArgs byte

func Command(client *rpc.Client) *ClientCommand {
	return &ClientCommand{client}
}

func (c *ClientCommand) Run(env packer.Environment, args []string) (result int) {
	// TODO: Environment
	rpcArgs := &CommandRunArgs{nil, args}
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
	*reply = c.command.Run(args.Env, args.Args)
	return nil
}

func (c *ServerCommand) Synopsis(args *CommandSynopsisArgs, reply *string) error {
	*reply = c.command.Synopsis()
	return nil
}
