package plugin

import (
	"bytes"
	"errors"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/rpc"
	"os/exec"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"strings"
	"time"
)

// Returns a valid packer.Command where the command is executed via RPC
// to a plugin that is within a subprocess.
//
// This method will start the given exec.Cmd, which should point to
// the plugin binary to execute. Some configuration will be done to
// the command, such as overriding Stdout and some environmental variables.
//
// This function guarantees the subprocess will end in a timely manner.
func Command(cmd *exec.Cmd) (result packer.Command, err error) {
	env := []string{
		"PACKER_PLUGIN_MIN_PORT=10000",
		"PACKER_PLUGIN_MAX_PORT=25000",
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err = cmd.Start()
	if err != nil {
		return
	}

	defer func() {
		// Make sure the command is properly killed in the case of an error
		if err != nil {
			cmd.Process.Kill()

			// Log the stderr, which should include any logs from the subprocess
			log.Print(stderr.String())
		}
	}()

	// Goroutine + channel to signal that the process exited
	cmdExited := make(chan bool)
	go func() {
		cmd.Wait()
		cmdExited <- true
	}()

	// Timer for a timeout
	cmdTimeout := time.After(1 * time.Minute)

	var address string
	for done := false; !done; {
		select {
		case <-cmdExited:
			err = errors.New("plugin exited before we could connect")
			done = true
		case <- cmdTimeout:
			err = errors.New("timeout while waiting for plugin to start")
			done = true
		default:
		}

		if line, lerr := stdout.ReadBytes('\n'); lerr == nil {
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
