package plugin

import (
	"bytes"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"os/exec"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"strings"
	"time"
)

type processCommand struct {
	cmd *exec.Cmd
}

func (c *processCommand) Run(e packer.Environment, args []string) int {
	return 0
}

func (c *processCommand) Synopsis() string {
	out := new(bytes.Buffer)
	c.cmd.Stdout = out
	c.cmd.Start()
	defer c.cmd.Process.Kill()

	// TODO: timeout
	// TODO: check that command is even running
	address := ""
	for {
		line, err := out.ReadBytes('\n')
		if err == nil {
			address = strings.TrimSpace(string(line))
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	client, _ := rpc.Dial("tcp", address)
	defer client.Close()

	realCommand := packrpc.Command(client)
	return realCommand.Synopsis()
}

func Command(cmd *exec.Cmd) packer.Command {
	return &processCommand{cmd}
}
