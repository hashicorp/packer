package packer

import "net/rpc"

// A command is a runnable sub-command of the `packer` application.
// When `packer` is called with the proper subcommand, this will be
// called.
//
// The mapping of command names to command interfaces is in the
// Environment struct.
//
// Run should run the actual command with the given environmet and
// command-line arguments. It should return the exit status when it is
// finished.
//
// Synopsis should return a one-line, short synopsis of the command.
// This should be less than 50 characters ideally.
type Command interface {
	Run(env Environment, args []string) int
	Synopsis() string
}

// An RPCCommand is an implementation of the Command interface where the
// command is actually executed over an RPC connection.
type RPCClientCommand struct {
	client *rpc.Client
}

// An RPCServerCommand wraps a Command and makes it exportable as part
// of a Golang RPC server.
type RPCServerCommand struct {
	command Command
}

type RPCCommandRunArgs struct {
	Env Environment
	Args []string
}

type RPCCommandSynopsisArgs byte

func (c *RPCClientCommand) Run(env Environment, args []string) (result int) {
	// TODO: Environment
	rpcArgs := &RPCCommandRunArgs{nil, args}
	c.client.Call("Command.Run", rpcArgs, &result)
	return
}

func (c *RPCClientCommand) Synopsis() (result string) {
	c.client.Call("Command.Synopsis", RPCCommandSynopsisArgs(0), &result)
	return
}

func (c *RPCServerCommand) Run(args *RPCCommandRunArgs, reply *int) error {
	*reply = c.command.Run(args.Env, args.Args)
	return nil
}

func (c *RPCServerCommand) Synopsis(args *RPCCommandSynopsisArgs, reply *string) error {
	*reply = c.command.Synopsis()
	return nil
}
