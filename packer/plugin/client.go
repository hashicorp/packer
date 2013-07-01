package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// This is a slice of the "managed" clients which are cleaned up when
// calling Cleanup
var managedClients = make([]*Client, 0, 5)

// Client handles the lifecycle of a plugin application, determining its
// RPC address, and returning various types of packer interface implementations
// across the multi-process communication layer.
type Client struct {
	config      *ClientConfig
	exited      bool
	doneLogging bool
	l           sync.Mutex
	address     string
}

// ClientConfig is the configuration used to initialize a new
// plugin client. After being used to initialize a plugin client,
// that configuration must not be modified again.
type ClientConfig struct {
	// The unstarted subprocess for starting the plugin.
	Cmd *exec.Cmd

	// Managed represents if the client should be managed by the
	// plugin package or not. If true, then by calling CleanupClients,
	// it will automatically be cleaned up. Otherwise, the client
	// user is fully responsible for making sure to Kill all plugin
	// clients. By default the client is _not_ managed.
	Managed bool

	// The minimum and maximum port to use for communicating with
	// the subprocess. If not set, this defaults to 10,000 and 25,000
	// respectively.
	MinPort, MaxPort uint

	// StartTimeout is the timeout to wait for the plugin to say it
	// has started successfully.
	StartTimeout time.Duration
}

// This makes sure all the managed subprocesses are killed and properly
// logged. This should be called before the parent process running the
// plugins exits.
//
// This must only be called _once_.
func CleanupClients() {
	// Kill all the managed clients in parallel and use a WaitGroup
	// to wait for them all to finish up.
	var wg sync.WaitGroup
	for _, client := range managedClients {
		wg.Add(1)

		go func(client *Client) {
			client.Kill()
			wg.Done()
		}(client)
	}

	log.Println("waiting for all plugin processes to complete...")
	wg.Wait()
}

// Creates a new plugin client which manages the lifecycle of an external
// plugin and gets the address for the RPC connection.
//
// The client must be cleaned up at some point by calling Kill(). If
// the client is a managed client (created with NewManagedClient) you
// can just call CleanupClients at the end of your program and they will
// be properly cleaned.
func NewClient(config *ClientConfig) (c *Client) {
	if config.MinPort == 0 && config.MaxPort == 0 {
		config.MinPort = 10000
		config.MaxPort = 25000
	}

	if config.StartTimeout == 0 {
		config.StartTimeout = 1 * time.Minute
	}

	c = &Client{config: config}
	if config.Managed {
		managedClients = append(managedClients, c)
	}

	return
}

// Tells whether or not the underlying process has exited.
func (c *Client) Exited() bool {
	return c.exited
}

// Returns a builder implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *Client) Builder() (packer.Builder, error) {
	client, err := c.rpcClient()
	if err != nil {
		return nil, err
	}

	return &cmdBuilder{packrpc.Builder(client), c}, nil
}

// Returns a command implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *Client) Command() (packer.Command, error) {
	client, err := c.rpcClient()
	if err != nil {
		return nil, err
	}

	return &cmdCommand{packrpc.Command(client), c}, nil
}

// Returns a hook implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *Client) Hook() (packer.Hook, error) {
	client, err := c.rpcClient()
	if err != nil {
		return nil, err
	}

	return &cmdHook{packrpc.Hook(client), c}, nil
}

// Returns a post-processor implementation that is communicating over
// this client. If the client hasn't been started, this will start it.
func (c *Client) PostProcessor() (packer.PostProcessor, error) {
	client, err := c.rpcClient()
	if err != nil {
		return nil, err
	}

	return &cmdPostProcessor{packrpc.PostProcessor(client), c}, nil
}

// Returns a provisioner implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *Client) Provisioner() (packer.Provisioner, error) {
	client, err := c.rpcClient()
	if err != nil {
		return nil, err
	}

	return &cmdProvisioner{packrpc.Provisioner(client), c}, nil
}

// End the executing subprocess (if it is running) and perform any cleanup
// tasks necessary such as capturing any remaining logs and so on.
//
// This method blocks until the process successfully exits.
//
// This method can safely be called multiple times.
func (c *Client) Kill() {
	cmd := c.config.Cmd

	if cmd.Process == nil {
		return
	}

	cmd.Process.Kill()

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

// Starts the underlying subprocess, communicating with it to negotiate
// a port for RPC connections, and returning the address to connect via RPC.
//
// This method is safe to call multiple times. Subsequent calls have no effect.
// Once a client has been started once, it cannot be started again, even if
// it was killed.
func (c *Client) Start() (address string, err error) {
	c.l.Lock()
	defer c.l.Unlock()

	if c.address != "" {
		return c.address, nil
	}

	env := []string{
		fmt.Sprintf("%s=%s", MagicCookieKey, MagicCookieValue),
		fmt.Sprintf("PACKER_PLUGIN_MIN_PORT=%d", c.config.MinPort),
		fmt.Sprintf("PACKER_PLUGIN_MAX_PORT=%d", c.config.MaxPort),
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := c.config.Cmd
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	log.Printf("Starting plugin: %s %#v", cmd.Path, cmd.Args)
	err = cmd.Start()
	if err != nil {
		return
	}

	// Make sure the command is properly cleaned up if there is an error
	defer func() {
		r := recover()

		if err != nil || r != nil {
			cmd.Process.Kill()
		}

		if r != nil {
			panic(r)
		}
	}()

	// Start goroutine to wait for process to exit
	go func() {
		cmd.Wait()
		log.Printf("%s: plugin process exited\n", cmd.Path)
		c.exited = true
	}()

	// Start goroutine that logs the stderr
	go c.logStderr(stderr)

	// Some channels for the next step
	timeout := time.After(c.config.StartTimeout)

	// Start looking for the address
	log.Printf("Waiting for RPC address for: %s", cmd.Path)
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
			c.address = strings.TrimSpace(string(line))
			address = c.address
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

func (c *Client) logStderr(buf *bytes.Buffer) {
	for done := false; !done; {
		if c.Exited() {
			done = true
		}

		var err error
		for err != io.EOF {
			var line string
			line, err = buf.ReadString('\n')
			if line != "" {
				log.Printf("%s: %s", c.config.Cmd.Path, line)
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Flag that we've completed logging for others
	c.doneLogging = true
}

func (c *Client) rpcClient() (*rpc.Client, error) {
	address, err := c.Start()
	if err != nil {
		return nil, err
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return client, nil
}
