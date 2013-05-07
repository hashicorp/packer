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

func Command(cmd *exec.Cmd) (result packer.Command, err error) {
	env := []string{
		"PACKER_PLUGIN_MIN_PORT=10000",
		"PACKER_PLUGIN_MAX_PORT=25000",
	}

	out := new(bytes.Buffer)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdout = out
	err = cmd.Start()
	if err != nil {
		return
	}

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

	result = packrpc.Command(client)
	return
}
