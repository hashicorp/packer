package plugin

import (
	"bufio"
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

type cmdCommand struct {
	command packer.Command
	exited <-chan bool
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
	select {
	case <-c.exited:
		cb()
	default:
		if p != nil {
			log.Panic(p)
		}
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
		}
	}()

	// Goroutine + channel to signal that the process exited
	cmdExited := make(chan bool)
	go func() {
		cmd.Wait()
		cmdExited <- true
	}()

	// Goroutine to log out the output from the command
	// TODO: All sorts of things wrong with this. First, we're reading from
	// a channel that can get consumed elsewhere. Second, the app can end
	// without properly flushing all the log data. BLah.
	go func() {
		buf := bufio.NewReader(stderr)

		for done := false; !done; {
			select {
			case <-cmdExited:
				done = true
			default:
			}

			var err error
			for err == nil {
				var line string
				line, err = buf.ReadString('\n')
				if line != "" {
					log.Print(line)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Timer for a timeout
	cmdTimeout := time.After(1 * time.Minute)

	var address string
	for done := false; !done; {
		select {
		case <-cmdExited:
			err = errors.New("plugin exited before we could connect")
			done = true
		case <-cmdTimeout:
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

	result = &cmdCommand{
		packrpc.Command(client),
		cmdExited,
	}

	return
}
