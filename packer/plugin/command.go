package plugin

import (
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"log"
	"net/rpc"
	"os/exec"
)

type cmdCommand struct {
	command packer.Command
	client  *client
}

func (c *cmdCommand) Run(e packer.Environment, args []string) (exitCode int) {
	defer func() {
		r := recover()
		c.checkExit(r, func() { exitCode = 1 })
	}()

	exitCode = c.command.Run(e, args)
	return
}

func (c *cmdCommand) Synopsis() (result string) {
	defer func() {
		r := recover()
		c.checkExit(r, func() {
			result = ""
		})
	}()

	result = c.command.Synopsis()
	return
}

func (c *cmdCommand) checkExit(p interface{}, cb func()) {
	if c.client.Exited() {
		cb()
	} else if p != nil {
		log.Panic(p)
	}
}

// Returns a valid packer.Command where the command is executed via RPC
// to a plugin that is within a subprocess.
//
// This method will start the given exec.Cmd, which should point to
// the plugin binary to execute. Some configuration will be done to
// the command, such as overriding Stdout and some environmental variables.
//
// This function guarantees the subprocess will end in a timely manner.
func Command(cmd *exec.Cmd) (result packer.Command, err error) {
	cmdClient := NewManagedClient(cmd)
	address, err := cmdClient.Start()
	if err != nil {
		return
	}

	defer func() {
		// Make sure the command is properly killed in the case of an error
		if err != nil {
			cmdClient.Kill()
		}
	}()

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return
	}

	result = &cmdCommand{
		packrpc.Command(client),
		cmdClient,
	}

	return
}
