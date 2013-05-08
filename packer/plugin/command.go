package plugin

import (
	"bytes"
	"errors"
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

	cmdExited := make(chan bool)
	go func() {
		cmd.Wait()
		cmdExited <- true
	}()

	var address string
	for done := false; !done; {
		select {
		case <-cmdExited:
			err = errors.New("plugin exited before we could connect")
			done = true
		default:
		}

		if line, lerr := out.ReadBytes('\n'); lerr == nil {
			// Trim the address and reset the err since we were able
			// to read some sort of address.
			address = strings.TrimSpace(string(line))
			err = nil
			break
		}

		// If error is nil from previously, return now
		if err != nil {
			return
		}

		// Wait a bit
		time.Sleep(10 * time.Millisecond)
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return
	}

	result = packrpc.Command(client)
	return
}
