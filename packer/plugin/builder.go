package plugin

import (
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"log"
	"net/rpc"
	"os/exec"
)

type cmdBuilder struct {
	builder packer.Builder
	client  *client
}

func (b *cmdBuilder) Prepare(config interface{}) error {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Prepare(config)
}

func (b *cmdBuilder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) packer.Artifact {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Run(ui, hook, cache)
}

func (b *cmdBuilder) Cancel() {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	b.builder.Cancel()
}

func (c *cmdBuilder) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil {
		log.Panic(p)
	}
}

// Returns a valid packer.Builder where the builder is executed via RPC
// to a plugin that is within a subprocess.
//
// This method will start the given exec.Cmd, which should point to
// the plugin binary to execute. Some configuration will be done to
// the command, such as overriding Stdout and some environmental variables.
//
// This function guarantees the subprocess will end in a timely manner.
func Builder(cmd *exec.Cmd) (result packer.Builder, err error) {
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

	result = &cmdBuilder{
		packrpc.Builder(client),
		cmdClient,
	}

	return
}
