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

func Command(cmd *exec.Cmd) packer.Command {
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Start()

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

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	return packrpc.Command(client)
}
