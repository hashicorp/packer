package plugin

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// This is a slice of the "managed" clients which are cleaned up when
// calling Cleanup
var managedClients = make([]*client, 0, 5)

type client struct {
	cmd *exec.Cmd
	exited bool
	doneLogging bool
}

// This makes sure all the managed subprocesses are killed and properly
// logged. This should be called before the parent process running the
// plugins exits.
func CleanupClients() {
	// Kill all the managed clients in parallel and use a WaitGroup
	// to wait for them all to finish up.
	var wg sync.WaitGroup
	for _, client := range managedClients {
		wg.Add(1)

		go func() {
			client.Kill()
			wg.Done()
		}()
	}

	log.Println("waiting for all plugin processes to complete...")
	wg.Wait()
}

func NewClient(cmd *exec.Cmd) *client {
	return &client{
		cmd,
		false,
		false,
	}
}

func NewManagedClient(cmd *exec.Cmd) (result *client) {
	result = NewClient(cmd)
	managedClients = append(managedClients, result)
	return
}

func (c *client) Exited() bool {
	return c.exited
}

func (c *client) Start() (address string, err error) {
	env := []string{
		"PACKER_PLUGIN_MIN_PORT=10000",
		"PACKER_PLUGIN_MAX_PORT=25000",
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	c.cmd.Env = append(c.cmd.Env, env...)
	c.cmd.Stderr = stderr
	c.cmd.Stdout = stdout
	err = c.cmd.Start()
	if err != nil {
		return
	}

	// Make sure the command is properly cleaned up if there is an error
	defer func() {
		r := recover()

		if err != nil || r != nil {
			c.cmd.Process.Kill()
		}

		if r != nil {
			panic(r)
		}
	}()

	// Start goroutine to wait for process to exit
	go func() {
		c.cmd.Wait()
		log.Println("plugin process exited")
		c.exited = true
	}()

	// Start goroutine that logs the stderr
	go c.logStderr(stderr)

	// Some channels for the next step
	timeout := time.After(1 * time.Minute)

	// Start looking for the address
	for done := false; !done; {
		select {
		case <-timeout:
			err = errors.New("timeout while waiting for plugin to start")
			done = true
		default:
		}

		if err == nil && c.Exited() {
			err = errors.New("plugin exited before we could connect")
			done = true
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

	return
}

func (c *client) Kill() {
	if c.cmd.Process == nil {
		return
	}

	c.cmd.Process.Kill()

	// Wait for the client to finish logging so we have a complete log
	done := make(chan bool)
	go func() {
		for !c.doneLogging {
			time.Sleep(10 * time.Millisecond)
		}

		done <- true
	}()

	<-done
}

func (c *client) logStderr(buf *bytes.Buffer) {
	for done := false; !done; {
		if c.Exited() {
			done = true
		}

		var err error
		for err != io.EOF {
			var line string
			line, err = buf.ReadString('\n')
			if line != "" {
				log.Printf("%s: %s", c.cmd.Path, line)
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Flag that we've completed logging for others
	c.doneLogging = true
}
