package plugin

import (
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"log"
	"net/rpc"
	"os/exec"
)

type cmdHook struct {
	hook packer.Hook
	client  *client
}

func (c *cmdHook) Run(name string, data interface{}, ui packer.Ui) {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	c.hook.Run(name, data, ui)
}

func (c *cmdHook) checkExit(p interface{}, cb func()) {
	if c.client.Exited() {
		cb()
	} else if p != nil {
		log.Panic(p)
	}
}

// Returns a valid packer.Hook where the hook is executed via RPC
// to a plugin that is within a subprocess.
//
// This method will start the given exec.Cmd, which should point to
// the plugin binary to execute. Some configuration will be done to
// the command, such as overriding Stdout and some environmental variables.
//
// This function guarantees the subprocess will end in a timely manner.
func Hook(cmd *exec.Cmd) (result packer.Hook, err error) {
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

	result = &cmdHook{
		packrpc.Hook(client),
		cmdClient,
	}

	return
}
